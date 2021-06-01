package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/pastdev/clconf/v2/clconf"
	"github.com/spf13/cobra"
)

type rootContext struct {
	ignoreEnv           bool
	prefix              optionalString
	secretKeyring       optionalString
	secretKeyringBase64 optionalString
	stdin               bool
	vars                []string
	yaml                []string
	yamlBase64          []string
}

func (c *rootContext) getPath(valuePath string) string {
	if c.prefix.set {
		return path.Join(c.prefix.value, valuePath)
	} else if prefix, ok := os.LookupEnv("CONFIG_PREFIX"); !c.ignoreEnv && ok {
		return path.Join(prefix, valuePath)
	}

	if valuePath == "" {
		return "/"
	}
	return valuePath
}

func (c *rootContext) getValue(path string) (interface{}, error) {
	path = c.getPath(path)

	confSources := clconf.ConfSources{
		Files:       c.yaml,
		Overrides:   c.yamlBase64,
		Environment: !c.ignoreEnv,
	}
	if c.stdin {
		confSources.Stream = os.Stdin
	}

	config, err := confSources.LoadInterface()
	if err != nil {
		return nil, err
	}
	if config == nil {
		config = map[interface{}]interface{}{}
	}

	for _, v := range c.vars {
		keyValue := strings.SplitN(v, "=", 2)
		if len(keyValue) != 2 {
			return nil, fmt.Errorf(
				"failed to parse var, expected `/key/path=\"jsonValue\"`, found: %s",
				v)
		}
		key := keyValue[0]

		value, err := clconf.UnmarshalSingleYaml(keyValue[1])
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal var %s: %v", key, err)
		}

		err = clconf.SetValue(config, key, value)
		if err != nil {
			return nil, fmt.Errorf("failed to set var %s: %v", key, err)
		}
	}

	return clconf.GetValue(config, path)
}

func rootCmd() *cobra.Command {
	var c = &rootContext{}
	pipe := false

	var cmd = &cobra.Command{
		Use: "clconf [global options] command [command options] [args...]",
		Short: `A utility for merging multiple config files and extracting values using a path
string`,
		Long: `A utility for merging multiple config files and extracting values using a path string
the order of precedence from least to greatest is:
  --yaml
  YAML_FILES
  --yaml-base64
  YAML_VARS
  --stdin`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if pipe {
				c.stdin = true
				c.ignoreEnv = true
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return (&getvContext{rootContext: c}).getv(cmd, args)
		},
		SilenceUsage: true,
	}

	cmd.PersistentFlags().BoolVar(
		&c.ignoreEnv,
		"ignore-env",
		false,
		"Tells clconf to use only command options (not environment variable equivalents).")
	cmd.PersistentFlags().Var(
		&c.prefix,
		"prefix",
		"Prepended to all getv/setv paths (env: CONFIG_PREFIX)")
	cmd.PersistentFlags().Var(
		&c.secretKeyring,
		"secret-keyring",
		"Path to a gpg secring file (env: SECRET_KEYRING)")
	cmd.PersistentFlags().Var(
		&c.secretKeyringBase64,
		"secret-keyring-base64",
		"Base64 encoded gpg secring (env: SECRET_KEYRING_BASE64)")
	cmd.PersistentFlags().BoolVarP(
		&pipe,
		"pipe",
		"p",
		false,
		"Shortcut for --stdin --ignore-env")
	cmd.PersistentFlags().BoolVar(
		&c.stdin,
		"stdin",
		false,
		"Read one or more yaml documents from stdin. Last document takes precedence when merged.")
	cmd.PersistentFlags().StringArrayVar(
		&c.vars,
		"var",
		nil,
		`A list of key=value pairs.  The key is a path into the config, and the value must be
yaml/json encoded.  Often combined with clconf var which produces well formed,
properly escaped vars (ie: clconf --var "$(clconf var /foo bar)")`)
	cmd.PersistentFlags().StringArrayVar(
		&c.yaml,
		"yaml",
		nil,
		`A list of yaml files containing config (env: YAML_FILES).  If specified, YAML_FILES will be
split on ',' and appended to this option.  Last defined value takes precedence when merged.`)
	cmd.PersistentFlags().StringArrayVar(
		&c.yamlBase64,
		"yaml-base64",
		nil,
		`A list of base 64 encoded yaml strings containing config (env: YAML_VARS).  If specified,
YAML_VARS will be split on ',' and each value will be used to load a base64 string from an
environtment variable of that name.  The values will be appended to this option.  Last defined value
takes precedence when merged`)

	cmd.AddCommand(
		cgetvCmd(c),
		csetvCmd(c),
		getvCmd(c),
		setvCmd(c),
		templateCmd(c),
		varCmd(),
		versionCmd())

	return cmd
}

// Execute runs the root command
func Execute() {
	if err := rootCmd().Execute(); err != nil {
		if e, ok := err.(*exitError); ok {
			os.Exit(e.exitCode)
		}
		os.Exit(1)
	}
}
