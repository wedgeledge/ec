/*
Copyright © 2022 Jonathan Holloway <jholloway@redhat.com>

*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// projectCmd represents the project command
var createProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("ERROR: must specify the name of the project")

			return
		}
		ecDir, _ := cmd.Flags().GetString("dir")
		ecConfig, _ := cmd.Flags().GetString("config")
		fmt.Println("create project called")
		fmt.Println("You're arguments were: " + strings.Join(args, ","))
		fmt.Println("Value of config flag: " + ecConfig)
		fmt.Println("Value of dir flag: " + ecDir)

		//project.RouteCLI(cmd, args)
	},
}

func init() {
	createCmd.AddCommand(createProjectCmd)
	//rootCmd.AddCommand(projectCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// projectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	createProjectCmd.Flags().StringP("dir", "d", "/tmp", "Specify a working directory for the project")
	createProjectCmd.Flags().StringP("config", "c", "~/.ec/config.yml", "Specify a config file for the project")

}
