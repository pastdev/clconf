package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/pastdev/clconf/clconf"
	"github.com/spf13/cobra"
)

var getvCmdContext = &getvContext{
	rootContext: rootCmdContext,
}

type getvContext struct {
	*rootContext
	decrypt        []string
	defaultValue   optionalString
	template       optionalString
	templateBase64 optionalString
	templateString optionalString
}

var getvCmd = &cobra.Command{
	Use:   "getv [options]",
	Short: "Get a value",
	RunE:  getv,
}

// Makes dump unit testable as test classes can override print
// https://stackoverflow.com/a/26804949/516433
var print = fmt.Print

func (c *getvContext) getPath(valuePath string) string {
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

func (c *getvContext) getTemplate() (*clconf.Template, error) {
	var tmpl *clconf.Template
	var err error
	if c.templateString.set {
		secretAgent, _ := c.newSecretAgent()
		tmpl, err = clconf.NewTemplate("cli", c.templateString.value,
			&clconf.TemplateConfig{
				SecretAgent: secretAgent,
			})
	}
	if err == nil && tmpl == nil {
		if c.templateBase64.set {
			secretAgent, _ := c.newSecretAgent()
			tmpl, err = clconf.NewTemplateFromBase64("cli", c.templateBase64.value,
				&clconf.TemplateConfig{
					SecretAgent: secretAgent,
				})
		}
	}
	if err == nil && tmpl == nil {
		if c.template.set {
			secretAgent, _ := c.newSecretAgent()
			tmpl, err = clconf.NewTemplateFromFile("cli", c.template.value,
				&clconf.TemplateConfig{
					SecretAgent: secretAgent,
				})
		}
	}
	return tmpl, err
}

func getv(cmd *cobra.Command, args []string) error {
	var path string
	if len(args) > 0 {
		path = args[0]
	}

	value, err := getvCmdContext.getValue(path)
	if err != nil {
		return err
	}

	value, err = getvCmdContext.marshal(value)
	if err != nil {
		return err
	}

	print(value)
	return nil
}

func (c *getvContext) getValue(path string) (interface{}, error) {
	path = c.getPath(path)
	config, err := clconf.LoadConfFromEnvironment(c.yaml, c.yamlBase64)
	if err != nil {
		return nil, err
	}

	value, ok := clconf.GetValue(config, path)
	if !ok {
		if c.defaultValue.set {
			value = c.defaultValue.value
		} else {
			return nil, NewExitError(1, fmt.Sprintf("[%v] does not exist", path))
		}
	}

	if len(c.decrypt) > 0 {
		secretAgent, err := c.newSecretAgent()
		if err != nil {
			return nil, err
		}
		if stringValue, ok := value.(string); ok {
			if len(c.decrypt) != 1 || !(c.decrypt[0] == "" || c.decrypt[0] == "/") {
				return nil, NewExitError(1, "string value with non-root decrypt path")
			}
			decrypted, err := secretAgent.Decrypt(stringValue)
			if err != nil {
				return nil, err
			}
			value = decrypted
		} else {
			err = secretAgent.DecryptPaths(value, c.decrypt...)
			if err != nil {
				return nil, err
			}
		}
	}
	return value, nil
}

func init() {
	rootCmd.AddCommand(getvCmd)

	getvCmd.Flags().StringArrayVarP(&getvCmdContext.decrypt, "decrypt", "", nil,
		"A `list` of paths whose values needs to be decrypted")
	getvCmd.Flags().VarP(&getvCmdContext.defaultValue, "default", "",
		"The value to be returned if the specified path does not exist (otherwise results in an error).")
	getvCmd.Flags().VarP(&getvCmdContext.template, "template", "",
		"A go template file that will be executed against the resulting data.")
	getvCmd.Flags().VarP(&getvCmdContext.templateBase64, "template-base64", "",
		"A base64 encoded string containing a go template that will be executed against the resulting data.")
	getvCmd.Flags().VarP(&getvCmdContext.templateString, "template-string", "",
		"A string containing a go template that will be executed against the resulting data.")
}

func (c *getvContext) marshal(value interface{}) (string, error) {
	if template, err := c.getTemplate(); err != nil {
		return "", err
	} else if template != nil {
		return template.Execute(value)
	} else if stringValue, ok := value.(string); ok {
		return stringValue, nil
	} else if mapValue, ok := value.(map[interface{}]interface{}); ok {
		if marshaled, err := clconf.MarshalYaml(mapValue); err != nil {
			return "", err
		} else {
			return string(marshaled), nil
		}
	} else if arrayValue, ok := value.([]interface{}); ok {
		if marshaled, err := clconf.MarshalYaml(arrayValue); err != nil {
			return "", err
		} else {
			return string(marshaled), nil
		}
	}
	return fmt.Sprintf("%v", value), nil
}

func (c *getvContext) newSecretAgent() (*clconf.SecretAgent, error) {
	var secretAgent *clconf.SecretAgent
	var err error

	if c.secretKeyringBase64.set {
		secretAgent, err = clconf.NewSecretAgentFromBase64(c.secretKeyringBase64.value)
	} else if c.secretKeyring.set {
		secretAgent, err = clconf.NewSecretAgentFromFile(c.secretKeyring.value)
	} else if keyBase64, ok := os.LookupEnv("SECRET_KEYRING_BASE64"); ok {
		secretAgent, err = clconf.NewSecretAgentFromBase64(keyBase64)
	} else if keyFile, ok := os.LookupEnv("SECRET_KEYRING"); ok {
		secretAgent, err = clconf.NewSecretAgentFromFile(keyFile)
	} else {
		err = errors.New("requires --secret-keyring-base64, --secret-keyring, or SECRET_KEYRING")
	}

	return secretAgent, err
}
