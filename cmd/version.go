package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "0.0.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display clconf version",
	Run:  func(command *cobra.Command, args []string) {
		fmt.Printf("Version: %s", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
