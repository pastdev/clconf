package cmd

import (
	"fmt"
	"path"
	"testing"

	"github.com/pastdev/clconf/v2/clconf"
)

func testMarshal(t *testing.T, message, expected string, data interface{}, marshaler Marshaler) {
	t.Run(message, func(t *testing.T) {
		actual, err := marshaler.Marshal(data)
		if err != nil {
			t.Fatalf("testMarshal %s execute failed: %s", message, err)
		}

		if !marshaler.pretty && !marshaler.template.set && !marshaler.templateBase64.set && !marshaler.templateString.set {
			assertYamlEqual(t, fmt.Sprintf("testMarshal (yaml) %s", message), expected, actual)
		} else {
			if expected != actual {
				t.Errorf("testMarshal %s: %s != %s", message, expected, actual)
			}
		}
	})
}

func TestMarshal(t *testing.T) {
	marshaler := Marshaler{}

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
	marshaler := Marshaler{}
	marshaler.asJSON = true

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

func TestMarshalJSONPretty(t *testing.T) {
	marshaler := Marshaler{}
	marshaler.asJSON = true
	marshaler.pretty = true

	testMarshal(t, "scalar", `"foo"`, "foo", marshaler)
	testMarshal(t, "scalar number", "1", 1, marshaler)
	testMarshal(t, "list",
		"[\n  \"bar\",\n  \"baz\"\n]",
		[]interface{}{"bar", "baz"},
		marshaler)
	testMarshal(t, "map",
		"{\n  \"foo\": \"bar\",\n  \"hip\": \"hop\"\n}",
		map[interface{}]interface{}{"foo": "bar", "hip": "hop"},
		marshaler)
	testMarshal(t, "map with list",
		"{\n  \"foo\": [\n    \"bar\",\n    \"baz\"\n  ]\n}",
		map[interface{}]interface{}{"foo": []interface{}{"bar", "baz"}},
		marshaler)
	testMarshal(t, "map with sub-map",
		"{\n  \"foo\": {\n    \"bar\": \"baz\"\n  }\n}",
		map[interface{}]interface{}{"foo": map[interface{}]interface{}{"bar": "baz"}},
		marshaler)
}

func TestMarshalKvJSON(t *testing.T) {
	marshaler := Marshaler{}
	marshaler.asKvJSON = true
	testMarshal(t, "kvjson map with list",
		"/foo/0: bar\n/foo/1: baz",
		map[interface{}]interface{}{"foo": []interface{}{"bar", "baz"}},
		marshaler)
	testMarshal(t, "kvjson map with sub-map",
		"/foo/bar: baz",
		map[interface{}]interface{}{"foo": map[interface{}]interface{}{"bar": "baz"}},
		marshaler)
}

func TestMarshalPrettyKvJSON(t *testing.T) {
	marshaler := Marshaler{}
	marshaler.pretty = true
	marshaler.asKvJSON = true
	testMarshal(t, "kvjson map with list",
		"{\n\"/foo/0\": \"bar\",\n\"/foo/1\": \"baz\"\n}",
		map[interface{}]interface{}{"foo": []interface{}{"bar", "baz"}},
		marshaler)
	testMarshal(t, "kvjson map with sub-map",
		"{\n\"/foo/bar\": \"baz\"\n}",
		map[interface{}]interface{}{"foo": map[interface{}]interface{}{"bar": "baz"}},
		marshaler)
}

func TestMarshalWithTemplate(t *testing.T) {
	keyFile := path.Join("..", "testdata", "test.secring.gpg")
	secretAgent, err := clconf.NewSecretAgentFromFile(keyFile)
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

	marshaler := Marshaler{
		secretAgentFactory: &rootContext{
			secretKeyring: *newOptionalString(keyFile, true),
		},
		templateString: *newOptionalString("{{getv \"/username\"}}:{{getv \"/password\"}}", true),
	}
	testMarshal(t, "basic template",
		"foo:bar",
		map[interface{}]interface{}{"username": "foo", "password": "bar"},
		marshaler)

	marshaler.templateString = *newOptionalString("{{cgetv \"/username\"}}:{{cgetv \"/password\"}}", true)
	testMarshal(t, "decrypt template",
		"foo:bar",
		map[interface{}]interface{}{"username": encryptedFoo, "password": encryptedBar},
		marshaler)
}
