package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type getvContext struct {
	*rootContext
	decrypt      []string
	defaultValue optionalString
	Marshaler
}

func (c *getvContext) addFlags(cmd *cobra.Command) {
	cmd.Flags().StringArrayVar(
		&c.decrypt,
		"decrypt",
		nil,
		"A `list` of paths whose values needs to be decrypted")
	cmd.Flags().Var(
		&c.defaultValue,
		"default",
		`The value to be returned if the specified path does not exist (otherwise results in an
error).`)
}

func (c *getvContext) getv(
	cmd *cobra.Command, //nolint:unparam
	args []string,
) error {
	var path string
	if len(args) > 0 {
		path = args[0]
	}

	value, err := c.getValue(path)
	if err != nil {
		return err
	}

	value, err = c.Marshal(value)
	if err != nil {
		return err
	}

	fmt.Print(value)
	return nil
}

func (c *getvContext) getValue(path string) (interface{}, error) {
	value, err := c.rootContext.getValue(path)
	if err != nil {
		if c.defaultValue.set {
			value = c.defaultValue.value
		} else {
			return nil, err
		}
	}

	if len(c.decrypt) > 0 {
		secretAgent, err := c.rootContext.newSecretAgent()
		if err != nil {
			return nil, err
		}
		if stringValue, ok := value.(string); ok {
			if len(c.decrypt) != 1 || !(c.decrypt[0] == "" || c.decrypt[0] == "/") {
				return nil, NewExitError(1, "string value with non-root decrypt path")
			}
			decrypted, err := secretAgent.Decrypt(stringValue)
			if err != nil {
				return nil, fmt.Errorf("decrypt: %w", err)
			}
			value = decrypted
		} else {
			err = secretAgent.DecryptPaths(value, c.decrypt...)
			if err != nil {
				return nil, fmt.Errorf("decrypt paths: %w", err)
			}
		}
	}
	return value, nil
}

func cgetvCmd(rootCmdContext *rootContext) *cobra.Command {
	var cmdContext = &getvContext{
		rootContext: rootCmdContext,
	}

	var cmd = &cobra.Command{
		Use:   "cgetv [options]",
		Short: "Get a secret value.  Simply an alias to `getv --decrypt /`",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdContext.decrypt = []string{"/"}
			return cmdContext.getv(cmd, args)
		},
	}

	cmdContext.addFlags(cmd)

	return cmd
}

func getvCmd(rootCmdContext *rootContext) *cobra.Command {
	var cmdContext = &getvContext{
		rootContext: rootCmdContext,
		Marshaler: Marshaler{
			secretAgentFactory: rootCmdContext,
		},
	}

	var cmd = &cobra.Command{
		Use:   "getv [options]",
		Short: "Get a value",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdContext.getv(cmd, args)
		},
	}

	cmdContext.addFlags(cmd)
	cmdContext.Marshaler.AddFlags(cmd)

	return cmd
}
