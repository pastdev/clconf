package cmd

import (
	"fmt"

	"github.com/pastdev/clconf/clconf"
	"github.com/spf13/cobra"
)

var setvCmdContext = &setvContext{
	rootContext: rootCmdContext,
}

type setvContext struct {
	*rootContext
	encrypt bool
}

var csetvCmd = &cobra.Command{
	Use:   "csetv key value [options]",
	Short: "Set PATH to the encrypted value of VALUE in the file indicated by the global option --yaml (must be single valued).  Simply an alias to `setv --encrypt`",
	RunE:  csetv,
}

var setvCmd = &cobra.Command{
	Use:   "setv key value [options]",
	Short: "Set PATH to VALUE in the file indicated by the global option --yaml (must be single valued).",
	Args:  cobra.ExactArgs(2),
	RunE:  setv,
}

func csetv(cmd *cobra.Command, args []string) error {
	setvCmdContext.encrypt = true
	return setv(cmd, args)
}

func init() {
	rootCmd.AddCommand(setvCmd, csetvCmd)

	setvCmd.Flags().BoolVarP(&setvCmdContext.encrypt, "encrypt", "", false,
		"Encrypt the value")

	csetvCmd.Flags().AddFlagSet(setvCmd.Flags())
}

func (c *setvContext) setValue(key, value string) error {
	path := c.getPath(key)
	file, config, err := clconf.LoadSettableConfFromEnvironment(c.yaml)
	if err != nil {
		return fmt.Errorf("Failed to load config %s: %s", c.yaml, err)
	}

	if c.encrypt {
		secretAgent, err := c.newSecretAgent()
		if err != nil {
			return fmt.Errorf("Failed to load secret agent: %s", err)
		}
		encrypted, err := secretAgent.Encrypt(value)
		if err != nil {
			return fmt.Errorf("Failed to encrypt: %s", err)
		}
		value = encrypted
	}

	err = clconf.SetValue(config, path, value)
	if err != nil {
		return fmt.Errorf("Failed to set vaule at %s: %s", path, err)
	}

	err = clconf.SaveConf(config, file)
	if err != nil {
		return fmt.Errorf("Failed to save config %s: %v", file, err)
	}

	return nil
}

func setv(cmd *cobra.Command, args []string) error {
	return setvCmdContext.setValue(args[0], args[1])
}
