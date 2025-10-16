package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of StartDB",
	Long:  `Print the version number and build information of StartDB.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("StartDB v1.0.0")
		fmt.Println("Build: Development")
		fmt.Println("Go Version: 1.21+")
	},
}
