package cmd

import (
	"fmt"
	"path"
	"testing"

	"github.com/pastdev/clconf/v3/pkg/secret"
)

type MarshalerTestCase struct {
	name      string
	marshaler Marshaler
}

func testMarshal(t *testing.T, message, expected string, data interface{}, marshalers ...MarshalerTestCase) {
	for _, c := range marshalers {
		t.Run(fmt.Sprintf("%s: %s", c.name, message), func(t *testing.T) {
			actual, err := c.marshaler.Marshal(data)
			if err != nil {
				t.Fatalf("testMarshal %s execute failed: %s", message, err)
			}

			if !c.marshaler.pretty && !c.marshaler.template.set && !c.marshaler.templateBase64.set && !c.marshaler.templateString.set {
				assertYamlEqual(t, fmt.Sprintf("testMarshal (yaml) %s", message), expected, actual)
			} else if expected != actual {
				t.Errorf("testMarshal %s: %v != %v", message, expected, actual)
			}
		})
	}
}

func TestMarshal(t *testing.T) {
	marshaler := MarshalerTestCase{name: "normal", marshaler: Marshaler{output: "yaml"}}

	testMarshal(t, "scalar", "foo", "foo", marshaler)
	testMarshal(t, "scalar number", "1", 1, marshaler)
	testMarshal(t, "list",
		"- bar\n- baz",
		[]interface{}{"bar", "baz"},
		marshaler)
	testMarshal(t, "map",
		"foo: bar\nhip: hop",
		map[interface{}]interface{}{"foo": "bar", "hip": "hop"},
		marshaler)
	testMarshal(t, "map with list",
		"foo:\n- bar\n- baz",
		map[interface{}]interface{}{"foo": []interface{}{"bar", "baz"}},
		marshaler)
	testMarshal(t, "map with sub-map",
		"foo:\n  bar:\n    baz",
		map[interface{}]interface{}{"foo": map[interface{}]interface{}{"bar": "baz"}},
		marshaler)
}

func TestMarshalJSON(t *testing.T) {
	marshalers := []MarshalerTestCase{
		{name: "normal", marshaler: Marshaler{output: "json"}},
		{name: "deprecated", marshaler: Marshaler{asJSON: true}},
	}

	testMarshal(t, "scalar", "foo", "foo", marshalers...)
	testMarshal(t, "scalar number", "1", 1, marshalers...)
	testMarshal(t, "list",
		"- bar\n- baz",
		[]interface{}{"bar", "baz"},
		marshalers...)
	testMarshal(t, "map",
		"foo: bar\nhip: hop",
		map[interface{}]interface{}{"foo": "bar", "hip": "hop"},
		marshalers...)
	testMarshal(t, "map with list",
		"foo:\n- bar\n- baz",
		map[interface{}]interface{}{"foo": []interface{}{"bar", "baz"}},
		marshalers...)
	testMarshal(t, "map with sub-map",
		"foo:\n  bar:\n    baz",
		map[interface{}]interface{}{"foo": map[interface{}]interface{}{"bar": "baz"}},
		marshalers...)
}

func TestMarshalJSONPretty(t *testing.T) {
	marshalers := []MarshalerTestCase{
		{name: "normal", marshaler: Marshaler{output: "json", pretty: true}},
		{name: "deprecated", marshaler: Marshaler{asJSON: true, pretty: true}},
	}

	testMarshal(t, "scalar", `"foo"`, "foo", marshalers...)
	testMarshal(t, "scalar number", "1", 1, marshalers...)
	testMarshal(t, "list",
		"[\n  \"bar\",\n  \"baz\"\n]",
		[]interface{}{"bar", "baz"},
		marshalers...)
	testMarshal(t, "map",
		"{\n  \"foo\": \"bar\",\n  \"hip\": \"hop\"\n}",
		map[interface{}]interface{}{"foo": "bar", "hip": "hop"},
		marshalers...)
	testMarshal(t, "map with list",
		"{\n  \"foo\": [\n    \"bar\",\n    \"baz\"\n  ]\n}",
		map[interface{}]interface{}{"foo": []interface{}{"bar", "baz"}},
		marshalers...)
	testMarshal(t, "map with sub-map",
		"{\n  \"foo\": {\n    \"bar\": \"baz\"\n  }\n}",
		map[interface{}]interface{}{"foo": map[interface{}]interface{}{"bar": "baz"}},
		marshalers...)
}

func TestMarshalKvJSON(t *testing.T) {
	marshalers := []MarshalerTestCase{
		{name: "normal", marshaler: Marshaler{output: "kv-json"}},
		{name: "deprecated", marshaler: Marshaler{asKvJSON: true}},
	}

	testMarshal(t, "kvjson map with list",
		"/foo/0: bar\n/foo/1: baz\n",
		map[interface{}]interface{}{"foo": []interface{}{"bar", "baz"}},
		marshalers...)
	testMarshal(t, "kvjson map with sub-map",
		"/foo/bar: baz\n",
		map[interface{}]interface{}{"foo": map[interface{}]interface{}{"bar": "baz"}},
		marshalers...)
}

func TestMarshalWithTemplate(t *testing.T) {
	keyFile := path.Join("..", "..", "testdata", "test.secring.gpg")
	secretAgent, err := secret.NewSecretAgentFromFile(keyFile)
	if err != nil {
		t.Errorf("Unable to load secret agent %s: %s", keyFile, err)
	}
	encryptedFoo, err := secretAgent.Encrypt("foo")
	if err != nil {
		t.Error("Unable to encrypt foo")
	}
	encryptedBar, err := secretAgent.Encrypt("bar")
	if err != nil {
		t.Error("Unable to encrypt bar")
	}

	marshalers := []MarshalerTestCase{
		{
			name: "normal",
			marshaler: Marshaler{
				output: "yaml",
				secretAgentFactory: &rootContext{
					secretKeyring: *newOptionalString(keyFile, true),
				},
				templateString: *newOptionalString("{{getv \"/username\"}}:{{getv \"/password\"}}", true),
			},
		},
	}
	testMarshal(t, "basic template",
		"foo:bar",
		map[interface{}]interface{}{"username": "foo", "password": "bar"},
		marshalers...)

	marshalers[0].marshaler.templateString = *newOptionalString("{{cgetv \"/username\"}}:{{cgetv \"/password\"}}", true)
	testMarshal(t, "decrypt template",
		"foo:bar",
		map[interface{}]interface{}{"username": encryptedFoo, "password": encryptedBar},
		marshalers...)
}
