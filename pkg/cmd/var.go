package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func varCmd() *cobra.Command {
	var forceArray bool
	var valueOnly bool

	cmd := &cobra.Command{
		Use:   "var",
		Short: "Print out a var in the format used with clconf --var",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			var marshaled []byte
			var err error
			if !forceArray && len(args) == 2 {
				marshaled, err = json.Marshal(args[1])
			} else {
				marshaled, err = json.Marshal(args[1:])
			}
			if err != nil {
				return fmt.Errorf("marshal var: %w", err)
			}

			if valueOnly {
				fmt.Printf("%s\n", marshaled)
			} else {
				fmt.Printf("%s=%s\n", args[0], marshaled)
			}
			return nil
		},
	}

	cmd.PersistentFlags().BoolVarP(
		&forceArray,
		"force-array",
		"a",
		false,
		"If true, an array of will be returned even if single valued.")
	cmd.PersistentFlags().BoolVarP(
		&valueOnly,
		"value-only",
		"v",
		false,
		"If true, only the value will be printed.")

	return cmd
}
