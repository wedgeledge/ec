package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
)

// ImagesAPI has just enough of the API to handle downloading the ISO
type ImagesAPI struct {
	Installer Installer `json:"Installer"`
}

// Installer is a substruct of ImagesAPI with the ISO info
type Installer struct {
	ImageBuildISOURL string `json:"ImageBuildISOURL"`
}

// DownloadProgress contains info for ISO download progress calculations
type DownloadProgress struct {
	BytesDownloaded int64
	FileSize        int64
}

// Write is called by the io.TeeReader for download progress calculations
func (dp *DownloadProgress) Write(partial []byte) (int, error) {
	current := len(partial)
	dp.BytesDownloaded += int64(current)
	dp.PrintProgress()
	return current, nil
}

// PrintProgress prints the current download information
func (dp DownloadProgress) PrintProgress() {
	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	percentComplete := float64(dp.BytesDownloaded) / float64(dp.FileSize) * 100
	fmt.Printf("\rDownloading... %v of %v (%.2f%%) complete           ",
		ByteSizeForHumans(dp.BytesDownloaded),
		ByteSizeForHumans(dp.FileSize),
		percentComplete)
}

func ByteSizeForHumans(bytes int64) string {
	const factor = 1000
	if bytes < factor {
		return fmt.Sprintf("%d B", bytes)
	}
	divisor := int64(factor)
	exponent := 0
	for i := bytes / factor; i >= factor; i /= factor {
		divisor *= factor
		exponent++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(divisor), "kMGTPE"[exponent])
}

// CallAPI makes a REST call against the Edge API
func CallAPI(method string, edgeURL string, fileHandle io.Reader) []byte {
	var req *http.Request
	var err error

	ProxyURL := os.Getenv("EDGE_PROXY_URL")
	//EdgeURL := os.Getenv("EDGE_URL")
	EdgeUsername := os.Getenv("EDGE_USERNAME")
	EdgePassword := os.Getenv("EDGE_PASSWORD")

	proxyURL, _ := url.Parse(ProxyURL)
	transport := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	client := &http.Client{Transport: transport}

	switch method {
	case "GET":
		req, err = http.NewRequest("GET", edgeURL, nil)
	case "POST":
		req, err = http.NewRequest("POST", edgeURL, fileHandle)
	}

	req.SetBasicAuth(EdgeUsername, EdgePassword)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get: %v\n", err)
	}
	body, err := ioutil.ReadAll(resp.Body)

	resp.Body.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "get: %v\n", err)
	}

	return body
	// TODO: return err too
}

// Image gets information for a specific image
func Image(ImageNumber string) {
	EdgeURL := "api/edge/v1/images/" + string(ImageNumber)
	body := CallAPI("GET", EdgeURL, nil)
	fmt.Printf("%s", body)
}

// ImageList gets a list of all images
func ImagesList() {
	EdgeURL := "api/edge/v1/images"
	body := CallAPI("GET", EdgeURL, nil)
	fmt.Printf("%s", body)
}

// ImageRepo gets the repo information for an image
func ImageRepo(ImageNumber string) {
	EdgeURL := "api/edge/v1/images/" + string(ImageNumber) + "/repo"
	body := CallAPI("GET", EdgeURL, nil)
	fmt.Printf("%s", body)
}

// ImageStatus gets the status of an image build
func ImageStatus(ImageNumber string) {
	EdgeURL := "api/edge/v1/images/" + string(ImageNumber) + "/status"
	body := CallAPI("GET", EdgeURL, nil)
	fmt.Printf("%s", body)
}

// ImageCreate requests a new Image Builder image
func ImageCreate(configFilename string) {
	EdgeURL := "api/edge/v1/images"
	//config, err := ioutil.ReadFile(configFilename)
	configFH, err := os.Open(configFilename)
	if err != nil {
		fmt.Printf("ERROR reading %s\n", configFilename)
		os.Exit(1)
	}

	//CallPostAPI(EdgeURL, string(config))
	body := CallAPI("POST", EdgeURL, configFH)
	configFH.Close()
	fmt.Printf("%s", body)
}

// ImageISO gets the ISO URL and optionally downloads the ISO
func ImageISO(imageNumber string, download bool, fqdir string, fqpath string, checksum bool) {
	EdgeURL := "api/edge/v1/images/" + string(imageNumber)
	body := CallAPI("GET", EdgeURL, nil)

	images := ImagesAPI{}
	json.Unmarshal(body, &images)
	fmt.Println(images.Installer.ImageBuildISOURL)

	if download {
		baseFilename := path.Base(images.Installer.ImageBuildISOURL)

		if fqpath == "" {
			if fqdir == "" {
				fqdir = "."
			}
			fqpath = fqdir + "/" + baseFilename
		}
		// TODO: check for dir path existence before trying to write

		isoURL := images.Installer.ImageBuildISOURL
		isoResponse, err := http.Head(isoURL)
		isoSize, _ := strconv.ParseInt(isoResponse.Header.Get("Content-Length"), 10, 64)
		fmt.Printf("Downloading %v bytes\n\tfrom %s\n\tto %s...\n", isoSize, isoURL, fqpath)

		isoResp, err := http.Get(isoURL)
		if err != nil {
			fmt.Println("ERROR: Cannot download file")
		}
		defer isoResp.Body.Close()

		localISOFile, err := os.Create(fqpath + ".partial")
		if err != nil {
			fmt.Println("ERROR: ISO file create")
		}

		progress := &DownloadProgress{}
		progress.FileSize = isoSize
		_, err = io.Copy(localISOFile, io.TeeReader(isoResp.Body, progress))

		fmt.Print("\n")
		localISOFile.Close()
		err = os.Rename(fqpath+".partial", fqpath)
		if err != nil {
			fmt.Println("ERROR renaming partial download")
		}

		// calculate checksum
		// TODO: compare to upstream file checksum
		if checksum {
			fmt.Println("Calculating sha256 checksum...")
			sumCalculator := sha256.New()
			sumfile, _ := ioutil.ReadFile(fqpath)
			sumCalculator.Write(sumfile)
			fmt.Printf("Checksum (sha256): %s\n", hex.EncodeToString(sumCalculator.Sum(nil)))
		}
	}
}

// OpenConsole opens the fleet management app webpage in the default browser
func OpenConsole(consoleURL string) {
	var err error

	switch runtime.GOOS {
	case "darwin":
		err = exec.Command("open", consoleURL).Start()
	case "linux":
		err = exec.Command("xdg-open", consoleURL).Start()
	default:
		fmt.Printf("ERROR opening %s: OS not recognized\n", consoleURL)
	}

	if err != nil {
		fmt.Printf("ERROR opening %s", consoleURL)
	}
}

// Main is main
func main() {

	/* SETUP THE ARGS */
	// images - list all
	imagesFS := flag.NewFlagSet("images", flag.ExitOnError)

	// image
	imageFS := flag.NewFlagSet("image", flag.ExitOnError)
	imageID := imageFS.String("id", "", "Image ID")

	// image-repo
	imageRepoFS := flag.NewFlagSet("image-repo", flag.ExitOnError)
	imageRepoID := imageRepoFS.String("id", "", "Image ID")

	// image-status
	imageStatusFS := flag.NewFlagSet("image-status", flag.ExitOnError)
	imageStatusID := imageStatusFS.String("id", "", "Image ID")

	// image-create
	imageCreateFS := flag.NewFlagSet("image-create", flag.ExitOnError)
	imageCreateConfig := imageCreateFS.String("f", "", "JSON file with create data")

	// image-iso
	imageISOFS := flag.NewFlagSet("image-iso", flag.ExitOnError)
	imageISOID := imageISOFS.String("id", "", "Image ID")
	imageISODownload := imageISOFS.Bool("download", false, "Download the iso?")
	imageISODir := imageISOFS.String("d", "", "Image output directory")
	imageISOOutfile := imageISOFS.String("o", "", "Image output filepath")
	imageISOSum := imageISOFS.Bool("checksum", false, "Calculate image checksum")

	// console
	consoleFS := flag.NewFlagSet("console", flag.ExitOnError)
	consoleStageEnv := consoleFS.Bool("stage", true, "Open stage console")

	if len(os.Args) < 2 {
		fmt.Println("Expected an edge command (e.g., edge images)")
		os.Exit(1)
	}

	/* HANDLE THE ARGS */
	switch os.Args[1] {
	case "image":
		imageFS.Parse(os.Args[2:])
		Image(*imageID)
	case "image-repo":
		imageRepoFS.Parse(os.Args[2:])
		ImageRepo(*imageRepoID)
	case "image-status":
		imageStatusFS.Parse(os.Args[2:])
		ImageStatus(*imageStatusID)
	case "image-create":
		imageCreateFS.Parse(os.Args[2:])
		ImageCreate(*imageCreateConfig)
	case "image-iso":
		imageISOFS.Parse(os.Args[2:])
		ImageISO(*imageISOID, *imageISODownload, *imageISODir, *imageISOOutfile, *imageISOSum)
	case "images":
		imagesFS.Parse(os.Args[2:])
		ImagesList()
	case "console":
		consoleURL := "https://$EDGE/edge/manage-images"
		consoleFS.Parse(os.Args[2:])
		// TODO: what is production console URL
		if *consoleStageEnv {
			consoleURL = "https://$EDGE/edge/manage-images"
		}
		OpenConsole(consoleURL)
	default:
		fmt.Println("Expected an edge command (e.g., edge images)")
		os.Exit(1)
	}
}

// TODO: move URL to env var (to allow pointing to stage or prod)
// TODO: json config file instead of or in addition to env vars
// TODO: add some error checking
// TODO: make proxy optional
// TODO: containerize this and stick it in Quay
// TODO: look at how oc handles Args (NOTE: It's using Cobra. Will port after adding image download)
// TODO: add checksum compare after download (requires edge-api update)
