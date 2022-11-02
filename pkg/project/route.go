package project

import (
	"fmt"
	"strings"

	"github.com/loadtheaccumulator/wedgeledge/ec/pkg/config"
	"github.com/spf13/cobra"
)

func RouteCLI(cmd *cobra.Command, args []string, config *config.EdgeConfig) {
	ecDir, _ := cmd.Flags().GetString("dir")
	ecConfig, _ := cmd.Flags().GetString("config")
	fmt.Println("project called")
	fmt.Println("Edge Username: ", config.EdgeUsername)
	fmt.Println("Called As:", cmd.CalledAs())
	fmt.Println("Command: ", cmd.CommandPath())
	fmt.Println("You're arguments were: " + strings.Join(args, ","))
	fmt.Println("(project route) Value of config flag: " + ecConfig)
	fmt.Println("Value of dir flag: " + ecDir)
}
