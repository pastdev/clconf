package yamljson_test

import (
	"reflect"
	"testing"

	"github.com/pastdev/clconf/v3/pkg/yamljson"
	"github.com/stretchr/testify/require"
)

const configMap = "" +
	"---\n" +
	"key: foobar\n" +
	"database:\n" +
	"  host: 127.0.0.1\n" +
	"  password: REDACTED\n" +
	"  port: \"3306\"\n" +
	"  username: REDACTED\n" +
	"upstream:\n" +
	"  app1: 10.0.1.10:8080\n" +
	"  app2: 10.0.1.11:8080\n" +
	"prefix:\n" +
	"  database:\n" +
	"    host: 127.0.0.1\n" +
	"    password: REDACTED\n" +
	"    port: \"3306\"\n" +
	"    username: REDACTED\n" +
	"  upstream:\n" +
	"    app1: 10.0.1.10:8080\n" +
	"    app2: 10.0.1.11:8080\n"
const secrets = "" +
	"---\n" +
	"database:\n" +
	"  password: p@sSw0rd\n" +
	"  username: confd\n" +
	"prefix:\n" +
	"  database:\n" +
	"    password: p@sSw0rd\n" +
	"    username: confd\n"

var configMapAndSecretsExpected = map[interface{}]interface{}{
	"key": "foobar",
	"database": map[interface{}]interface{}{
		"host":     "127.0.0.1",
		"password": "p@sSw0rd",
		"port":     "3306",
		"username": "confd",
	},
	"upstream": map[interface{}]interface{}{
		"app1": "10.0.1.10:8080",
		"app2": "10.0.1.11:8080",
	},
	"prefix": map[interface{}]interface{}{
		"database": map[interface{}]interface{}{
			"host":     "127.0.0.1",
			"password": "p@sSw0rd",
			"port":     "3306",
			"username": "confd",
		},
		"upstream": map[interface{}]interface{}{
			"app1": "10.0.1.10:8080",
			"app2": "10.0.1.11:8080",
		},
	},
}

func TestUnmarshalSingleYaml(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		yamlObj, err := yamljson.UnmarshalSingleYaml("---")
		if err != nil {
			t.Errorf("failed to unmarshal; %v", err)
		}
		if yamlObj != nil {
			t.Errorf("expected <nil> for empty `---`, got %T: %v", yamlObj, yamlObj)
		}
	})
	t.Run("string", func(t *testing.T) {
		actual, err := yamljson.UnmarshalSingleYaml("foo")
		if err != nil {
			t.Errorf("failed to unmarshal; %v", err)
		}
		expected := "foo"
		if expected != actual {
			t.Errorf("%v != %v", expected, actual)
		}
	})
	t.Run("number", func(t *testing.T) {
		actual, err := yamljson.UnmarshalSingleYaml("10")
		if err != nil {
			t.Errorf("failed to unmarshal; %v", err)
		}
		expected := 10
		if expected != actual {
			t.Errorf("%v != %v", expected, actual)
		}
	})
	t.Run("array", func(t *testing.T) {
		yamlObj, err := yamljson.UnmarshalSingleYaml("[\"bar\", \"baz\"]")
		if err != nil {
			t.Errorf("failed to unmarshal; %v", err)
		}
		if actual, ok := yamlObj.([]interface{}); !ok {
			t.Errorf("%v not an slice", yamlObj)
		} else {
			expected := []interface{}{"bar", "baz"}
			if !reflect.DeepEqual(expected, actual) {
				t.Errorf("%v != %v", expected, actual)
			}
		}
	})
}

func TestUnmarshalYaml(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		expected := map[interface{}]interface{}{}
		actual, err := yamljson.UnmarshalYamlInterface("---")
		if err != nil {
			t.Fatalf("Error UnmarshalYaml: %v", err)
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Empty failed: [%#v] != [%#v]", expected, actual)
		}
	})
	t.Run("configMapAndSecrets", func(t *testing.T) {
		actual, err := yamljson.UnmarshalYamlInterface(configMap, secrets)
		require.NoError(t, err)
		require.Equal(t, configMapAndSecretsExpected, actual)
		// if err != nil || !reflect.DeepEqual(actual, configMapAndSecretsExpected) {
		// 	t.Errorf("ConfigMap and Secrets failed: [%v] != [%v]", configMapAndSecretsExpected, actual)
		// }
	})
}

func TestUnmarshalYamlInterface(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		expected := map[interface{}]interface{}{}
		actual, err := yamljson.UnmarshalYamlInterface("---")
		if err != nil {
			t.Fatalf("Error UnmarshalYaml: %v", err)
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Empty failed: [%#v] != [%#v]", expected, actual)
		}
	})
	t.Run("configMapAndSecrets", func(t *testing.T) {
		actual, err := yamljson.UnmarshalYamlInterface(configMap, secrets)
		if err != nil || !reflect.DeepEqual(actual, configMapAndSecretsExpected) {
			t.Errorf("ConfigMap and Secrets failed: [%#v] != [%#v]", configMapAndSecretsExpected, actual)
		}
	})
}

func TestUnmarshalYamlMultipleDocs(t *testing.T) {
	expected := map[interface{}]interface{}{
		"a": "foo",
		"b": "bar",
		"c": "stuff",
	}

	merged, err := yamljson.UnmarshalYamlInterface("---\na: bar\n---\nb: bar", "---\na: foo\n---\nc: stuff\n")
	if err != nil || !reflect.DeepEqual(merged, expected) {
		t.Errorf("Multiple docs failed: [%v] != [%v]", expected, merged)
	}
}

func TestUnmarshalYamlInterfaceMultipleDocs(t *testing.T) {
	t.Run("hashes", func(t *testing.T) {
		expected := map[interface{}]interface{}{
			"a": "foo",
			"b": "bar",
			"c": "stuff",
		}
		merged, err := yamljson.UnmarshalYamlInterface("---\na: bar\n---\nb: bar", "---\na: foo\n---\nc: stuff\n")
		if err != nil || !reflect.DeepEqual(merged, expected) {
			t.Errorf("Multiple docs failed: [%#v] != [%#v]", expected, merged)
		}
	})
	t.Run("hashes then array", func(t *testing.T) {
		expected := []interface{}{
			"one",
			"two",
		}
		merged, err := yamljson.UnmarshalYamlInterface("---\na: bar\n---\nb: bar", "---\na: foo\n---\nc: stuff\n---\n- one\n- two")
		if err != nil || !reflect.DeepEqual(merged, expected) {
			t.Errorf("Multiple docs failed: [%#v] != [%#v]", expected, merged)
		}
	})
	t.Run("hashes then nil", func(t *testing.T) {
		expected := map[interface{}]interface{}{
			"a": "foo",
			"b": "bar",
			"c": "stuff",
		}
		merged, err := yamljson.UnmarshalYamlInterface("---\na: bar\n---\nb: bar", "---\na: foo\n---\nc: stuff\n---")
		if err != nil || !reflect.DeepEqual(merged, expected) {
			t.Errorf("Multiple docs failed: [%#v] != [%#v]", expected, merged)
		}
	})
}

func TestUnmarshalYamlJson(t *testing.T) {
	expected := map[interface{}]interface{}{
		"a": "bar",
	}

	merged, err := yamljson.UnmarshalYamlInterface(`{"a": "bar",}`)
	if err != nil || !reflect.DeepEqual(merged, expected) {
		t.Errorf("Json failed: [%v] != [%v]: %v", expected, merged, err)
	}
}

func TestUnmarshalYamlMultipleJsons(t *testing.T) {
	expected := map[interface{}]interface{}{
		"a": "bar",
		"b": "foo",
	}

	merged, err := yamljson.UnmarshalYamlInterface("{\"a\": \"bar\",}\n---\n{\"b\": \"foo\"}")
	if err != nil || !reflect.DeepEqual(merged, expected) {
		t.Errorf("Multiple json failed: [%v] != [%v]: %v", expected, merged, err)
	}
}

func TestUnmarshalYamlNumericKey(t *testing.T) {
	expected := map[interface{}]interface{}{
		1: "one",
		2: "two",
	}
	actual, err := yamljson.UnmarshalYamlInterface("1: one\n2: two")
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("Numeric keys failed: [%v] != [%v]", expected, actual)
	}
}
