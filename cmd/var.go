package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func varCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "var",
		Short: "Print out a var in the format used with clconf --var",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			var marshaled []byte
			var err error
			if len(args) == 2 {
				marshaled, err = json.Marshal(args[1])
			} else {
				marshaled, err = json.Marshal(args[1:])
			}
			if err != nil {
				return fmt.Errorf("unable to marshal var: %v", err)
			}

			fmt.Printf("%s=%s\n", args[0], marshaled)
			return nil
		},
	}
}
