package cmd

import (
	"fmt"

	"github.com/ohler55/ojg/jp"
	"github.com/spf13/cobra"
)

type jsonpathContext struct {
	*rootContext
	first bool
	Marshaler
}

func (c *jsonpathContext) addFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(
		&c.first,
		"first",
		"0",
		false,
		"Prints the first element in the result array")
}

func (c jsonpathContext) jsonpath(cmd *cobra.Command, args []string) error {
	path := "$"
	if len(args) > 0 {
		path = args[0]
	}

	x, err := jp.ParseString(path)
	if err != nil {
		return err
	}
	data, err := c.rootContext.getValue("/")
	if err != nil {
		return err
	}

	var value interface{}
	result := x.Get(convertMapIToMapS(data))
	value = result
	if c.first {
		value = result[0]
	}

	marshalled, err := c.Marshal(value)
	if err != nil {
		return err
	}

	fmt.Print(marshalled)
	return nil
}

func jsonpathCmd(rootCmdContext *rootContext) *cobra.Command {
	var cmdContext = &jsonpathContext{
		rootContext: rootCmdContext,
		Marshaler: Marshaler{
			secretAgentFactory: rootCmdContext,
		},
	}

	var cmd = &cobra.Command{
		Use:   "jsonpath [options]",
		Short: "Get the value at the supplied path",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdContext.jsonpath(cmd, args)
		},
	}

	cmdContext.addFlags(cmd)
	cmdContext.Marshaler.AddFlags(cmd)

	return cmd
}
