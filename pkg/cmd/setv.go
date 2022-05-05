package cmd

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/pastdev/clconf/v2/pkg/conf"
	"github.com/pastdev/clconf/v2/pkg/core"
	"github.com/pastdev/clconf/v2/pkg/yamljson"
	"github.com/spf13/cobra"
)

type setvContext struct {
	*rootContext
	base64Value    bool
	encrypt        bool
	merge          bool
	mergeOverwrite bool
	yamlValue      bool
}

func csetvCmd(rootCmdContext *rootContext) *cobra.Command {
	var cmdContext = &setvContext{
		rootContext: rootCmdContext,
	}

	var cmd = &cobra.Command{
		Use:   "csetv key value [options]",
		Short: "Set PATH to the encrypted value of VALUE in the file indicated by the global option --yaml (must be single valued).  Simply an alias to `setv --encrypt`",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdContext.encrypt = true
			return cmdContext.setValue(args[0], args[1])
		},
	}

	cmdContext.addFlags(cmd)

	return cmd
}

func setvCmd(rootCmdContext *rootContext) *cobra.Command {
	var cmdContext = &setvContext{
		rootContext: rootCmdContext,
	}

	var cmd = &cobra.Command{
		Use:   "setv key value [options]",
		Short: "Set PATH to VALUE in the file indicated by the global option --yaml (must be single valued).",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdContext.setValue(args[0], args[1])
		},
	}

	cmdContext.addFlags(cmd)

	return cmd
}

func (c *setvContext) addFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&c.encrypt, "encrypt", "", false,
		"Encrypt the value")
	cmd.Flags().BoolVarP(&c.base64Value, "base64-value", "", false,
		"The value is base64 encoded")
	cmd.Flags().BoolVarP(&c.mergeOverwrite, "merge-overwrite", "", false,
		"Merged values should overwrite existing values (not used unless --merge)")
	cmd.Flags().BoolVarP(&c.merge, "merge", "", false,
		"Values should be merged rather than overwrite at path")
	cmd.Flags().BoolVarP(&c.yamlValue, "yaml-value", "", false,
		"The value is yaml/json")
}

func (c *setvContext) setValue(key, value string) error {
	if c.encrypt && c.yamlValue {
		return errors.New("--encrypt only works on strings (mutually exclusive with --yaml-value)")
	}

	path := c.getPath(key)
	config, file, err := conf.
		ConfSources{Environment: true, Files: c.yaml}.
		LoadSettableInterface()
	if err != nil {
		return fmt.Errorf("load config %s: %w", c.yaml, err)
	}

	if c.base64Value {
		newValue, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return fmt.Errorf("base64 decode %s: %w", value, err)
		}
		value = string(newValue)
	}

	if c.encrypt {
		secretAgent, err := c.newSecretAgent()
		if err != nil {
			return fmt.Errorf("load secret agent: %w", err)
		}
		encrypted, err := secretAgent.Encrypt(value)
		if err != nil {
			return fmt.Errorf("encrypt: %w", err)
		}
		value = encrypted
	}

	var valueObject interface{}
	valueObject = value

	if c.yamlValue {
		newValue, err := yamljson.UnmarshalSingleYaml(value)
		if err != nil {
			return fmt.Errorf("unmarshal %s: %w", value, err)
		}
		valueObject = newValue
	}

	if c.merge {
		err = core.MergeValue(config, path, valueObject, c.mergeOverwrite)
		if err != nil {
			return fmt.Errorf("merge value at %s: %w", path, err)
		}
	} else {
		err = core.SetValue(config, path, valueObject)
		if err != nil {
			return fmt.Errorf("set value at %s: %w", path, err)
		}
	}

	err = core.SaveConf(config, file)
	if err != nil {
		return fmt.Errorf("save config %s: %w", file, err)
	}

	return nil
}
