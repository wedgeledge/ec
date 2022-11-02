package image

import (
	"fmt"

	"github.com/spf13/cobra"
)

// RouteCLI handles and routes a call from the ec CLI tool
func RouteCLI(cmd *cobra.Command, args []string) {
	fmt.Println("image called")

	//List()
}
