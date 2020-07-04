package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/pastdev/clconf/v2/clconf"
	"github.com/spf13/cobra"
)

type getvContext struct {
	*rootContext
	asJSON         bool
	asKvJSON       bool
	decrypt        []string
	defaultValue   optionalString
	pretty         bool
	template       optionalString
	templateBase64 optionalString
	templateString optionalString
}

func (c *getvContext) addFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(
		&c.asJSON,
		"as-json",
		false,
		"Prints value as json")
	cmd.Flags().BoolVar(
		&c.asKvJSON,
		"as-kv-json",
		false,
		"Prints value as json formatted key/value pairs")
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
	cmd.Flags().BoolVar(
		&c.pretty,
		"pretty",
		false,
		"Pretty prints json output")
	cmd.Flags().Var(
		&c.template,
		"template",
		"A go template file that will be executed against the resulting data.")
	cmd.Flags().Var(
		&c.templateBase64,
		"template-base64",
		`A base64 encoded string containing a go template that will be executed against the
resulting data.`)
	cmd.Flags().Var(
		&c.templateString,
		"template-string",
		"A string containing a go template that will be executed against the resulting data.")
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

func (c *getvContext) getv(cmd *cobra.Command, args []string) error {
	var path string
	if len(args) > 0 {
		path = args[0]
	}

	value, err := c.getValue(path)
	if err != nil {
		return err
	}

	value, err = c.marshal(value)
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

func (c *getvContext) marshal(value interface{}) (string, error) {
	template, err := c.getTemplate()
	if err != nil {
		return "", err
	}

	if template != nil {
		return template.Execute(value)
	}

	if stringValue, ok := value.(string); ok {
		if c.asJSON {
			marshaled, _ := json.Marshal(stringValue)
			return string(marshaled), nil
		}
		return stringValue, nil
	}

	_, mapOk := value.(map[interface{}]interface{})
	_, arrayOk := value.([]interface{})

	if mapOk || arrayOk {
		var marshaled []byte
		var err error
		if c.asJSON {
			if c.pretty {
				marshaled, err = json.MarshalIndent(convertMapIToMapS(value), "", "  ")
			} else {
				marshaled, err = json.Marshal(convertMapIToMapS(value))
			}
		} else if c.asKvJSON {
			if c.pretty {
				marshaled, err = json.MarshalIndent(clconf.ToKvMap(value), "", "")
			} else {
				marshaled, err = json.Marshal(clconf.ToKvMap(value))
			}
		} else {
			marshaled, err = clconf.MarshalYaml(value)
		}
		if err != nil {
			return "", err
		}
		return string(marshaled), nil
	}

	return fmt.Sprintf("%v", value), nil
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

// json serialization requires string keys, but yaml deserialization creates
// interface{} keys.  This function converts the keys to strings so json can
// marshal the value
//   https://stackoverflow.com/a/40737676/516433
func convertMapIToMapS(mapI interface{}) interface{} {
	switch x := mapI.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convertMapIToMapS(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convertMapIToMapS(v)
		}
	}
	return mapI
}

func getvCmd(rootCmdContext *rootContext) *cobra.Command {
	var cmdContext = &getvContext{
		rootContext: rootCmdContext,
	}

	var cmd = &cobra.Command{
		Use:   "getv [options]",
		Short: "Get a value",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdContext.getv(cmd, args)
		},
	}

	cmdContext.addFlags(cmd)

	return cmd
}
