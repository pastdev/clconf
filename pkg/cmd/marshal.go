package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/pastdev/clconf/v3/pkg/core"
	"github.com/pastdev/clconf/v3/pkg/secret"
	"github.com/pastdev/clconf/v3/pkg/template"
	"github.com/pastdev/clconf/v3/pkg/yamljson"
	"github.com/spf13/cobra"
)

type secretAgentFactory interface {
	newSecretAgent() (*secret.SecretAgent, error)
}

type Marshaler struct {
	asBashArray bool
	asJSON      bool
	asJSONLines bool
	asKvJSON    bool
	leftDelim   string
	pretty      bool
	rightDelim  string
	secretAgentFactory
	template       optionalString
	templateBase64 optionalString
	templateString optionalString
	Executor
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
		&c.asJSONLines,
		"as-json-lines",
		false,
		"Prints each top level element as its own json structure on a single line")
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
	c.Executor.AddFlags(cmd)
}

func (c Marshaler) newSecretAgent() (*secret.SecretAgent, error) {
	if c.secretAgentFactory == nil {
		return nil, nil
	}
	a, err := c.secretAgentFactory.newSecretAgent()
	if err != nil {
		return nil, fmt.Errorf("new secret agent: %w", err)
	}
	return a, nil
}

func (c Marshaler) getTemplate() (*template.Template, error) {
	var tmpl *template.Template
	var err error
	secretAgent, _ := c.newSecretAgent()
	config := &template.TemplateConfig{
		SecretAgent: secretAgent,
		LeftDelim:   c.leftDelim,
		RightDelim:  c.rightDelim,
	}
	if c.templateString.set {
		tmpl, err = template.NewTemplate("cli", c.templateString.value, config)
	}
	if err == nil && tmpl == nil {
		if c.templateBase64.set {
			tmpl, err = template.NewTemplateFromBase64("cli", c.templateBase64.value, config)
		}
	}
	if err == nil && tmpl == nil {
		if c.template.set {
			tmpl, err = template.NewTemplateFromFile("cli", c.template.value, config)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("new template: %w", err)
	}
	return tmpl, nil
}

func (c Marshaler) Marshal(value interface{}) (string, error) {
	if len(c.execs) > 0 {
		secretAgent, _ := c.newSecretAgent()
		return "", c.Execute(value, &template.TemplateConfig{
			SecretAgent: secretAgent,
			LeftDelim:   c.leftDelim,
			RightDelim:  c.rightDelim,
		})
	}
	if c.asBashArray {
		return c.marshalBashArray(value)
	}
	if c.asJSONLines {
		return c.marshalJSONLines(value)
	}

	template, err := c.getTemplate()
	if err != nil {
		return "", err
	}

	if template != nil {
		t, err := template.Execute(value)
		if err != nil {
			return "", fmt.Errorf("template execute: %w", err)
		}
		return t, nil
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
		switch {
		case c.asJSON:
			if c.pretty {
				marshaled, err = json.MarshalIndent(yamljson.ConvertMapIToMapS(value), "", "  ")
			} else {
				marshaled, err = json.Marshal(yamljson.ConvertMapIToMapS(value))
			}
		case c.asKvJSON:
			if c.pretty {
				marshaled, err = json.MarshalIndent(core.ToKvMap(value), "", "")
			} else {
				marshaled, err = json.Marshal(core.ToKvMap(value))
			}
		default:
			marshaled, err = yamljson.MarshalYaml(value)
		}
		if err != nil {
			return "", fmt.Errorf("marshal: %w", err)
		}
		return string(marshaled), nil
	}

	return fmt.Sprintf("%v", value), nil
}

func (c Marshaler) marshalJSONLines(value interface{}) (string, error) {
	lines, err := ToJSONLines(value)
	if err != nil {
		return "", err
	}
	return strings.Join(lines, "\n"), nil
}

func (c Marshaler) marshalBashArray(value interface{}) (string, error) {
	lines, err := ToJSONLines(value)
	if err != nil {
		return "", err
	}

	first := true
	var builder strings.Builder
	builder.WriteString("(")
	for i, val := range lines {
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

func ToJSONLines(value interface{}) ([]string, error) {
	marshaler := Marshaler{asJSON: true}

	values := []string{}
	switch v := value.(type) {
	case []interface{}:
		for i, val := range v {
			val, err := marshaler.Marshal(val)
			if err != nil {
				return nil, fmt.Errorf("marshal value at %d: %w", i, err)
			}
			values = append(values, val)
		}
	case map[interface{}]interface{}:
		keys := []string{}
		for key, val := range v {
			keys = append(keys, fmt.Sprintf("%s", key))
			val, err := marshaler.Marshal(map[interface{}]interface{}{"key": key, "value": val})
			if err != nil {
				return nil, fmt.Errorf("marshal value at %s: %w", key, err)
			}
			values = append(values, val)
		}
		sort.Slice(values, func(i, j int) bool { return keys[i] < keys[j] })
	default:
		val, err := marshaler.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("marshal value: %w", err)
		}
		values = append(values, val)
	}

	return values, nil
}
