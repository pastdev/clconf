package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/pastdev/clconf/v2/clconf"
	"github.com/spf13/cobra"
)

type secretAgentFactory interface {
	newSecretAgent() (*clconf.SecretAgent, error)
}

type Marshaler struct {
	asBashArray bool
	asJSON      bool
	asKvJSON    bool
	leftDelim   string
	pretty      bool
	rightDelim  string
	secretAgentFactory
	template       optionalString
	templateBase64 optionalString
	templateString optionalString
}

func (c *Marshaler) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(
		&c.asBashArray,
		"as-bash-array",
		false,
		"Prints value in format compatable with bash `declare -a`")
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
	cmd.Flags().StringVar(
		&c.leftDelim,
		"left-delimiter",
		"{{",
		"Delimiter to use when parsing templates for substitutions")
	cmd.Flags().StringVar(
		&c.rightDelim,
		"right-delimiter",
		"}}",
		"Delimiter to use when parsing templates for substitutions")
}

func (c Marshaler) newSecretAgent() (*clconf.SecretAgent, error) {
	if c.secretAgentFactory == nil {
		return nil, nil
	}
	return c.secretAgentFactory.newSecretAgent()
}

func (c Marshaler) getTemplate() (*clconf.Template, error) {
	var tmpl *clconf.Template
	var err error
	secretAgent, _ := c.newSecretAgent()
	config := &clconf.TemplateConfig{
		SecretAgent: secretAgent,
		LeftDelim:   c.leftDelim,
		RightDelim:  c.rightDelim,
	}
	if c.templateString.set {
		tmpl, err = clconf.NewTemplate("cli", c.templateString.value, config)
	}
	if err == nil && tmpl == nil {
		if c.templateBase64.set {
			tmpl, err = clconf.NewTemplateFromBase64("cli", c.templateBase64.value, config)
		}
	}
	if err == nil && tmpl == nil {
		if c.template.set {
			tmpl, err = clconf.NewTemplateFromFile("cli", c.template.value, config)
		}
	}
	return tmpl, err
}

func (c Marshaler) Marshal(value interface{}) (string, error) {
	if c.asBashArray {
		return c.marshalBashArray(value)
	}

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

func (c Marshaler) marshalBashArray(value interface{}) (string, error) {
	c.asJSON = true
	c.asBashArray = false

	values := []string{}
	switch v := value.(type) {
	case []interface{}:
		for i, val := range v {
			val, err := c.Marshal(val)
			if err != nil {
				return "", fmt.Errorf("unable to marshal value at %d: %v", i, err)
			}
			values = append(values, val)
		}
	case map[interface{}]interface{}:
		keys := []string{}
		for key, val := range v {
			keys = append(keys, fmt.Sprintf("%s", key))
			val, err := c.Marshal(map[interface{}]interface{}{"key": key, "value": val})
			if err != nil {
				return "", fmt.Errorf("unable to marshal value at %s: %v", key, err)
			}
			values = append(values, val)
		}
		sort.Slice(values, func(i, j int) bool { return keys[i] < keys[j] })
	default:
		val, err := c.Marshal(value)
		if err != nil {
			return "", fmt.Errorf("unable to marshal value: %v", err)
		}
		values = append(values, val)
	}

	first := true
	var builder strings.Builder
	builder.WriteString("(")
	for i, val := range values {
		if first {
			first = false
		} else {
			builder.WriteString(" ")
		}
		builder.WriteString(fmt.Sprintf("[%d]=%s", i, bashEscape(val)))
	}
	builder.WriteString(")")

	return builder.String(), nil
}

func bashEscape(s string) string {
	// if already quoted, nothing to do
	if s[0] == '"' {
		return s
	}

	var builder strings.Builder
	builder.WriteRune('"')
	for _, r := range s {
		if r == '"' {
			builder.WriteRune('\\')
		}
		builder.WriteRune(r)
	}
	builder.WriteRune('"')

	return builder.String()
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
