package api

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/loadtheaccumulator/wedgeledge/ec/pkg/config"
)

//var ecfg = config.Get("")

func JoinURL(cfg *config.EdgeConfig, endpoint string) string {
	url, err := url.Parse(cfg.EdgeBaseURL)
	if err != nil {
		fmt.Println("Cannot parse EdgeBaseURL")
	}
	url.Path = path.Join(url.Path, endpoint)

	return url.String()
}

func Call(cfg *config.EdgeConfig, method string, url string, fh io.Reader) []byte {
	edgeAPI := &API{
		Method: method,
		URL:    url,
		FH:     fh,
	}

	body := edgeAPI.call(cfg)

	return body
}

// Call makes a REST call against the Edge API
func (api *API) call(cfg *config.EdgeConfig) []byte {
	var req *http.Request
	var err error
	var client *http.Client

	client = &http.Client{}

	if cfg.EdgeProxy.URL != "" {
		ProxyURL := cfg.EdgeProxy.URL

		proxyURL, _ := url.Parse(ProxyURL)
		transport := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
		client = &http.Client{Transport: transport}
	}

	switch api.Method {
	case "GET":
		req, err = http.NewRequest("GET", api.URL, nil)
	case "POST":
		req, err = http.NewRequest("POST", api.URL, api.FH)
	}

	req.SetBasicAuth(cfg.EdgeUsername, cfg.EdgePassword)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return nil
	}
	body, err := ioutil.ReadAll(resp.Body)

	resp.Body.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "get: %v\n", err)
	}

	return body
	// TODO: return err too
}
