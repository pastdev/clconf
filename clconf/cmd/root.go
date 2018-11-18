package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "clconf [global options] command [command options] [args...]",
	Short: `A utility for merging multiple config files and extracting values using a path string`,
	RunE:  getv,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
