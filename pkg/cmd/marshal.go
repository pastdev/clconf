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

const DefaultOutput = "value"

type secretAgentFactory interface {
	newSecretAgent() (*secret.SecretAgent, error)
}

var marshalOutputOptions = map[string]string{
	"bash-array":         "print the config as a string formatted for deserialization using bash declare -a",
	"go-template":        "process the config through a go template supplied via --template",
	"go-template-file":   "process the config through a go template supplied via a file named by --template",
	"go-template-base64": "process the config through a base64 encoded go template supplied via a --template",
	"json":               "print the config as a single line json object",
	"json-lines":         "print each top level element in the config as a single line json object. if the top level is a map, the json lines object will have two top level elements: `key` and `value`",
	"kv-json":            "print the config as a single line json object after mapping to key/value pairs",
	"yaml":               "print the config in yaml format",
	"value":              "like yaml, except that if it is a scalar value, it will not be quoted. this is the legacy format and thus set as default for backwards compatibility reasons",
}

type Marshaler struct {
	// hidden options kept for backwards compatibility
	asBashArray bool
	asJSON      bool
	asKvJSON    bool
	secretAgentFactory
	templateBase64 optionalString
	templateString optionalString

	// output sets how the configuration object will be rendered. the supported
	// output options are defined by marshalOutputOptions along with their
	// specification
	output string
	// leftDelim is the left delimiter used by the template engine for
	// placeholders
	leftDelim string
	// pretty will cause the output to be printed in a human readable format
	// if the selected --output options supports it
	pretty bool
	// rightDelim is the right delimiter used by the template engine for
	// placeholders
	rightDelim string
	// template defines the template to be used to process the configuration
	// object through. the meaing of this value depends on the value of output.
	template optionalString
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
	cmd.Flags().Var(
		&c.templateBase64,
		"template-base64",
		`A base64 encoded string containing a go template that will be executed against the resulting data.`)
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
	deprecateFlag(cmd, "as-bash-array")
	deprecateFlag(cmd, "as-json")
	deprecateFlag(cmd, "as-kv-json")
	deprecateFlag(cmd, "left-delimiter")
	deprecateFlag(cmd, "right-delimiter")
	deprecateFlag(cmd, "template-base64")
	deprecateFlag(cmd, "template-string")

	cmd.Flags().StringVar(
		&c.output,
		"output",
		DefaultOutput,
		usageForMarshalOutput())
	cmd.Flags().BoolVar(
		&c.pretty,
		"pretty",
		false,
		"Pretty prints output when possible")
	cmd.Flags().Var(
		&c.template,
		"template",
		"A `template` that will be used to process the resulting data. For backwards compatibility, if `--output` is not explicitly specified but `--template` is, then `--output go-template-file` is assumed")
	cmd.Flags().StringVar(
		&c.leftDelim,
		"template-left",
		"{{",
		"Delimiter to use when parsing templates for substitutions")
	cmd.Flags().StringVar(
		&c.rightDelim,
		"template-right",
		"}}",
		"Delimiter to use when parsing templates for substitutions")
}

func (c Marshaler) getTemplate() (*template.Template, error) {
	var tmpl *template.Template
	var err error
	config := c.getTemplateConfig()
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

func (c Marshaler) getTemplateConfig() *template.TemplateConfig {
	secretAgent, _ := c.newSecretAgent()
	return &template.TemplateConfig{
		SecretAgent: secretAgent,
		LeftDelim:   c.leftDelim,
		RightDelim:  c.rightDelim,
	}
}

func (c Marshaler) Marshal(value interface{}) (string, error) {
	switch {
	case c.output == "bash-array" || c.asBashArray:
		return marshalBashArray(value)
	case c.output == "json-lines":
		return marshalJSONLines(value)
	case c.output == "go-template" || c.templateString.set:
		var t string
		if c.output == "go-template" {
			t = c.template.value
		} else {
			t = c.templateString.value
		}
		tmpl, err := template.NewTemplate("cli", t, c.getTemplateConfig())
		if err != nil {
			return "", fmt.Errorf("new template from string %s: %w", c.template.value, err)
		}
		return marshalTemplate(tmpl, value)
	case c.output == "go-template-base64" || c.templateBase64.set:
		var t string
		if c.output == "go-template-base64" {
			t = c.template.value
		} else {
			t = c.templateBase64.value
		}
		tmpl, err := template.NewTemplateFromBase64("cli", t, c.getTemplateConfig())
		if err != nil {
			return "", fmt.Errorf("new template from base64 %s: %w", c.template.value, err)
		}
		return marshalTemplate(tmpl, value)
	case c.output == "go-template-file" || (c.template.set && c.output == DefaultOutput):
		tmpl, err := template.NewTemplateFromFile("cli", c.template.value, c.getTemplateConfig())
		if err != nil {
			return "", fmt.Errorf("new template from file %s: %w", c.template.value, err)
		}
		return marshalTemplate(tmpl, value)
	case c.output == "json" || c.asJSON:
		return marshalJSON(yamljson.ConvertMapIToMapS(value), c.pretty)
	case c.output == "kv-json" || c.asKvJSON:
		return marshalJSON(core.ToKvMap(value), c.pretty)
	case c.output == "yaml":
		return marshalYaml(value)
	case c.output == "value":
		if stringValue, ok := value.(string); ok {
			return stringValue, nil
		}
		return marshalYaml(value)
	default:
		return "", fmt.Errorf("output %s is not supported", c.output)
	}
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

func deprecateFlag(cmd *cobra.Command, name string) {
	// use MarkHidden instead of deprecate so as to not leak the deprecation
	// message into the stderr of current consumers for safer backwards compat.
	err := cmd.Flags().MarkHidden(name)
	if err != nil {
		panic(fmt.Sprintf("failed to deprectate %s", name))
	}
}

func marshalBashArray(value interface{}) (string, error) {
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

func marshalJSON(value interface{}, pretty bool) (string, error) {
	if stringValue, isString := value.(string); isString {
		marshaled, _ := json.Marshal(stringValue)
		return string(marshaled), nil
	}

	_, isMap := value.(map[string]interface{})
	_, isKvMap := value.(map[string]string)
	_, isArray := value.([]interface{})
	if isMap || isKvMap || isArray {
		var marshaled []byte
		var err error
		if pretty {
			marshaled, err = json.MarshalIndent(value, "", "  ")
		} else {
			marshaled, err = json.Marshal(value)
		}
		if err != nil {
			return "", fmt.Errorf("marshal: %w", err)
		}
		return string(marshaled), nil
	}

	return fmt.Sprintf("%v", value), nil
}

func marshalJSONLines(value interface{}) (string, error) {
	lines, err := ToJSONLines(value)
	if err != nil {
		return "", err
	}
	return strings.Join(lines, "\n"), nil
}

func marshalTemplate(tmpl *template.Template, value interface{}) (string, error) {
	t, err := tmpl.Execute(value)
	if err != nil {
		return "", fmt.Errorf("template execute: %w", err)
	}
	return t, nil
}

func marshalYaml(value interface{}) (string, error) {
	marshaled, err := yamljson.MarshalYaml(value)
	if err != nil {
		return "", fmt.Errorf("marshal yaml: %w", err)
	}
	return string(marshaled), nil
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

func usageForMarshalOutput() string {
	keys := make([]string, len(marshalOutputOptions))
	i := 0
	for k := range marshalOutputOptions {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	var builder strings.Builder
	builder.WriteString("One of the following `option`s:\n")
	for _, k := range keys {
		builder.WriteString("  * ")
		builder.WriteString(k)
		builder.WriteString(": ")
		builder.WriteString(marshalOutputOptions[k])
		builder.WriteString("\n")
	}
	builder.WriteString("See also `--template*`.")

	return builder.String()
}
