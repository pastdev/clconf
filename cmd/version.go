package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "0.0.0"

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Display clconf version",
		Run: func(command *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", version)
		},
	}
}
