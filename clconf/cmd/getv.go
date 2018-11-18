package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/urfave/cli"
)

var getvCmd = &cobra.Command{
	Use:   "getv [options]",
	Short: "Get a value",
	RunE:  getv,
}

// Makes dump unit testable as test classes can override print
// https://stackoverflow.com/a/26804949/516433
var print = fmt.Print

func dump(c *cli.Context, value interface{}, err cli.ExitCoder) cli.ExitCoder {
	if err != nil {
		return err
	}
	print(value)
	return nil
}

func getTemplate(c *cli.Context) (*cli.Context, *Template, cli.ExitCoder) {
	var tmpl *Template
	var err error
	templateString := c.String("template-string")
	if templateString != "" {
		secretAgent, _ := newSecretAgentFromCli(c)
		tmpl, err = NewTemplate("cli", templateString,
			&TemplateConfig{
				SecretAgent: secretAgent,
			})
	}
	if err == nil && tmpl == nil {
		templateBase64 := c.String("template-base64")
		if templateBase64 != "" {
			secretAgent, _ := newSecretAgentFromCli(c)
			tmpl, err = NewTemplateFromBase64("cli", templateBase64,
				&TemplateConfig{
					SecretAgent: secretAgent,
				})
		}
	}
	if err == nil && tmpl == nil {
		templateFile := c.String("template")
		if templateFile != "" {
			secretAgent, _ := newSecretAgentFromCli(c)
			tmpl, err = NewTemplateFromFile("cli", templateFile,
				&TemplateConfig{
					SecretAgent: secretAgent,
				})
		}
	}
	return c, tmpl, cliError(err, 1)
}

func getv(cmd *cobra.Command, args []string) error {
	return dump(marshal(getValue(c)))
}

func getValue(c *cli.Context) (*cli.Context, interface{}, cli.ExitCoder) {
	path := getPath(c)
	config, err := load(c)
	if err != nil {
		return c, nil, cliError(err, 1)
	}
	value, ok := GetValue(config, path)
	if !ok {
		value, ok = getDefault(c)
		if !ok {
			return c, nil, cli.NewExitError(fmt.Sprintf("[%v] does not exist", path), 1)
		}
	}
	if decryptPaths := c.StringSlice("decrypt"); len(decryptPaths) > 0 {
		secretAgent, err := newSecretAgentFromCli(c)
		if err != nil {
			return c, nil, err
		}
		if stringValue, ok := value.(string); ok {
			if len(decryptPaths) != 1 || !(decryptPaths[0] == "" || decryptPaths[0] == "/") {
				return c, nil, cli.NewExitError("string value with non-root decrypt path", 1)
			}
			decrypted, err := secretAgent.Decrypt(stringValue)
			if err != nil {
				return c, nil, cliError(err, 1)
			}
			value = decrypted
		} else {
			err = cliError(secretAgent.DecryptPaths(value, decryptPaths...), 1)
			if err != nil {
				return c, nil, err
			}
		}
	}
	return c, value, nil
}

func init() {
	rootCmd.AddCommand(getvCmd)
}

func marshal(c *cli.Context, value interface{}, err cli.ExitCoder) (*cli.Context, string, cli.ExitCoder) {
	if err != nil {
		return c, "", err
	}
	if _, tmpl, err := getTemplate(c); err != nil {
		return c, "", err
	} else if tmpl != nil {
		marshaled, err := tmpl.Execute(value)
		return c, marshaled, cliError(err, 1)
	} else if stringValue, ok := value.(string); ok {
		return c, stringValue, nil
	} else if mapValue, ok := value.(map[interface{}]interface{}); ok {
		marshaled, err := MarshalYaml(mapValue)
		return c, string(marshaled), cliError(err, 1)
	} else if arrayValue, ok := value.([]interface{}); ok {
		marshaled, err := MarshalYaml(arrayValue)
		return c, string(marshaled), cliError(err, 1)
	}
	return c, fmt.Sprintf("%v", value), err
}
