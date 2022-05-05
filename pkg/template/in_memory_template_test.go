package template_test

import (
	"testing"

	"github.com/pastdev/clconf/v2/pkg/secret"
	"github.com/pastdev/clconf/v2/pkg/template"
)

func TestNewTemplate(t *testing.T) {
	_, err := template.NewTemplate("foo", "bar", nil)
	if err != nil {
		t.Errorf("NewTemplate without config failed: %v", err)
	}

	_, err = template.NewTemplate("foo", "bar", &template.TemplateConfig{})
	if err != nil {
		t.Errorf("NewTemplate with config failed: %v", err)
	}
}

func testExecute(t *testing.T, name, text, prefix string, data interface{}, expected string) {
	sa, err := secret.NewTestSecretAgent()
	if err != nil {
		t.Errorf("Execute %s create secret agent failed: %v", name, err)
	}

	tmpl, err := template.NewTemplate(name, text,
		&template.TemplateConfig{
			Prefix:      prefix,
			SecretAgent: sa,
		})
	if err != nil {
		t.Errorf("Execute %s create template failed: %v", name, err)
	}
	result, err := tmpl.Execute(data)
	if err != nil {
		t.Errorf("Execute %s empty data failed: %v", name, err)
	}
	if result != expected {
		t.Errorf("Execute %s empty data invalid: [%v] != [%v]", name, result, expected)
	}
}

func TestExecute(t *testing.T) {
	testExecute(t, "empty template", "", "", nil, "")
	testExecute(t, "empty data", "foo", "", nil, "foo")
	testExecute(t, "no placeholder",
		"foo", "",
		map[interface{}]interface{}{"foo": "bar"},
		"foo")
	testExecute(t, "simple placeholder",
		"foo{{ getv \"/foo\" }}", "",
		map[interface{}]interface{}{"foo": "bar"},
		"foobar")
	testExecute(t, "simple placeholder with prefix",
		"foo{{ getv \"/bar\" }}", "/foo",
		map[interface{}]interface{}{
			"foo": map[interface{}]interface{}{
				"bar": "baz",
			},
		},
		"foobaz")
}
