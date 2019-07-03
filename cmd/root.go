package cmd

import (
	"os"
	"path"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "clconf [global options] command [command options] [args...]",
	Short: `A utility for merging multiple config files and extracting values using a path string`,
	RunE:  getv,
}

var rootCmdContext = &rootContext{}

type rootContext struct {
	prefix              optionalString
	secretKeyring       optionalString
	secretKeyringBase64 optionalString
	yaml                []string
	yamlBase64          []string
}

func (c *rootContext) getPath(valuePath string) string {
	if c.prefix.set {
		return path.Join(c.prefix.value, valuePath)
	} else if prefix, ok := os.LookupEnv("CONFIG_PREFIX"); ok {
		return path.Join(prefix, valuePath)
	}

	if valuePath == "" {
		return "/"
	}
	return valuePath
}

func init() {
	rootCmd.PersistentFlags().VarP(&rootCmdContext.prefix, "prefix", "",
		"Prepended to all getv/setv paths (env: CONFIG_PREFIX)")
	rootCmd.PersistentFlags().VarP(&rootCmdContext.secretKeyring, "secret-keyring", "",
		"Path to a gpg secring file (env: SECRET_KEYRING)")
	rootCmd.PersistentFlags().VarP(&rootCmdContext.secretKeyringBase64, "secret-keyring-base64", "",
		"Base64 encoded gpg secring (env: SECRET_KEYRING_BASE64)")
	rootCmd.PersistentFlags().StringArrayVarP(&rootCmdContext.yaml, "yaml", "", nil,
		"A `list` of yaml files containing config (env: YAML_FILES).  If specified, YAML_FILES will be split on ',' and appended to this option.  Last defined value takes precedence when merged.")
	rootCmd.PersistentFlags().StringArrayVarP(&rootCmdContext.yamlBase64, "yaml-base64", "", nil,
		"A `list` of base 64 encoded yaml strings containing config (env: YAML_VARS).  If specified, YAML_VARS will be split on ',' and each value will be used to load a base64 string from an environtment variable of that name.  The values will be appended to this option.  Last defined value takes precedence when merged")
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if e, ok := err.(*exitError); ok {
			os.Exit(e.exitCode)
		}
		os.Exit(1)
	}
}
