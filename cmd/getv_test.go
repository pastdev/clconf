package cmd

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/pastdev/clconf/v2/clconf"
)

func testBase64Yaml(yaml string) []string {
	return []string{base64.StdEncoding.EncodeToString([]byte(yaml))}
}

func testGetPath(t *testing.T, expected, path string, prefix optionalString) {
	testRootContext := &rootContext{prefix: prefix}
	context := &getvContext{rootContext: testRootContext}
	actual := context.getPath(path)
	if expected != actual {
		t.Errorf("Get path failed: [%v] != [%v]", expected, actual)
	}
}

func TestGetPath(t *testing.T) {
	var envVar = "CONFIG_PREFIX"
	defer func() {
		os.Unsetenv(envVar)
	}()

	testGetPath(t, "/", "", *newOptionalString("", false))
	testGetPath(t, "/", "/", *newOptionalString("", false))
	testGetPath(t, "/foo", "/foo", *newOptionalString("", false))

	testGetPath(t, "/foo", "", *newOptionalString("/foo", true))
	testGetPath(t, "/foo", "/", *newOptionalString("/foo", true))
	testGetPath(t, "/foo/bar", "/bar", *newOptionalString("/foo", true))
	testGetPath(t, "/foo/bar", "/bar", *newOptionalString("/foo/", true))

	os.Setenv(envVar, "/foo")
	testGetPath(t, "/foo", "", *newOptionalString("", false))
	testGetPath(t, "/foo", "/", *newOptionalString("", false))
	testGetPath(t, "/foo/bar", "/bar", *newOptionalString("", false))
}

func testGetValue(t *testing.T, message string, expected interface{}, path string, context getvContext) {
	actual, err := context.getValue(path)
	if err != nil {
		t.Errorf("testGetValue %s failed to get value: [%s]", message, path)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("testGetValue %s: %v != %v", message, expected, actual)
	}
}

func testGetvContext(yaml string) getvContext {
	return getvContext{
		rootContext: &rootContext{
			yamlBase64: testBase64Yaml(yaml),
		},
	}
}

func TestGetValue(t *testing.T) {
	testGetValue(t, "empty path",
		map[interface{}]interface{}{"foo": "bar"},
		"",
		testGetvContext("foo: bar"))
	testGetValue(t, "/ path",
		map[interface{}]interface{}{"foo": "bar"},
		"/",
		testGetvContext("foo: bar"))
	testGetValue(t, "/foo path",
		"bar",
		"/foo",
		testGetvContext("foo: bar"))
	testGetValue(t, "array value",
		[]interface{}{"foo", "bar"},
		"/array",
		testGetvContext("array:\n- foo\n- bar"))

	context := testGetvContext("foo: bar")
	context.defaultValue = *newOptionalString("baz", true)
	testGetValue(t, "default existing",
		"bar",
		"/foo",
		context)
	testGetValue(t, "default not existing",
		"baz",
		"/hip",
		context)

	keyFile := path.Join("..", "testdata", "test.secring.gpg")
	secretAgent, err := clconf.NewSecretAgentFromFile(keyFile)
	if err != nil {
		t.Errorf("Unable to load secret agent %s: %s", keyFile, err)
	}
	encryptedBar, err := secretAgent.Encrypt("bar")
	if err != nil {
		t.Error("Unable to encrypt bar")
	}
	encryptedHop, err := secretAgent.Encrypt("hop")
	if err != nil {
		t.Error("Unable to encrypt hop")
	}

	context = testGetvContext(fmt.Sprintf("foo: %s", encryptedBar))
	context.secretKeyring = *newOptionalString(keyFile, true)
	context.decrypt = []string{"/"}
	testGetValue(t, "decrypt value", "bar", "/foo", context)

	context = testGetvContext(
		fmt.Sprintf("foo: %s\nhip: %s\ntik: tok",
			encryptedBar, encryptedHop))
	context.secretKeyring = *newOptionalString(keyFile, true)
	context.decrypt = []string{"/foo", "/hip"}
	testGetValue(t, "decrypt multiple paths",
		map[interface{}]interface{}{"foo": "bar", "hip": "hop", "tik": "tok"},
		"/",
		context)
}

func testGetTemplate(t *testing.T, message string, expected interface{}, data interface{}, context getvContext) {
	template, err := context.getTemplate()
	if err != nil {
		t.Errorf("testGetTemplate %s failed: %s", message, err)
	}

	actual, err := template.Execute(data)
	if err != nil {
		t.Errorf("testGetTemplate %s execute failed: %s", message, err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("testGetTemplate %s: %v != %v", message, expected, actual)
	}
}

func TestGetTemplate(t *testing.T) {
	keyFile := path.Join("..", "testdata", "test.secring.gpg")

	tempDir, err := ioutil.TempDir("", "clconf")
	if err != nil {
		t.Errorf("Unable to create temp dir: %v", err)
	}
	defer func() {
		os.RemoveAll(tempDir)
	}()

	data := map[interface{}]interface{}{"foo": "bar"}
	templateString := "{{ getv \"/foo\" }}"
	templateBytes := []byte(templateString)
	templateBase64 := base64.StdEncoding.EncodeToString(templateBytes)
	templateFile := filepath.Join(tempDir, "template")
	ioutil.WriteFile(templateFile, templateBytes, 0600)

	context := getvContext{
		rootContext: &rootContext{
			secretKeyring: *newOptionalString(keyFile, true),
		},
		templateString: *newOptionalString(templateString, true),
	}
	testGetTemplate(t, "template string", "bar", data, context)

	context = getvContext{
		rootContext: &rootContext{
			secretKeyring: *newOptionalString(keyFile, true),
		},
		templateBase64: *newOptionalString(templateBase64, true),
	}
	testGetTemplate(t, "template base64", "bar", data, context)

	context = getvContext{
		rootContext: &rootContext{
			secretKeyring: *newOptionalString(keyFile, true),
		},
		template: *newOptionalString(templateFile, true),
	}
	testGetTemplate(t, "template file", "bar", data, context)
}

func testMarshal(t *testing.T, message, expected string, data interface{}, context getvContext) {
	actual, err := context.marshal(data)
	if err != nil {
		t.Errorf("testMarshal %s execute failed: %s", message, err)
	}

	if !context.pretty && !context.template.set && !context.templateBase64.set && !context.templateString.set {
		assertYamlEqual(t, fmt.Sprintf("testMarshal (yaml) %s", message), expected, actual)
	} else {
		if expected != actual {
			t.Errorf("testMarshal %s: %s != %s", message, expected, actual)
		}
	}
}

func TestMarshal(t *testing.T) {
	context := getvContext{}

	testMarshal(t, "scalar", "foo", "foo", context)
	testMarshal(t, "scalar number", "1", 1, context)
	testMarshal(t, "list",
		"- bar\n- baz",
		[]interface{}{"bar", "baz"},
		context)
	testMarshal(t, "map",
		"foo: bar\nhip: hop",
		map[interface{}]interface{}{"foo": "bar", "hip": "hop"},
		context)
	testMarshal(t, "map with list",
		"foo:\n- bar\n- baz",
		map[interface{}]interface{}{"foo": []interface{}{"bar", "baz"}},
		context)
	testMarshal(t, "map with sub-map",
		"foo:\n  bar:\n    baz",
		map[interface{}]interface{}{"foo": map[interface{}]interface{}{"bar": "baz"}},
		context)
}

func TestMarshalJSON(t *testing.T) {
	context := getvContext{}
	context.asJSON = true

	testMarshal(t, "scalar", "foo", "foo", context)
	testMarshal(t, "scalar number", "1", 1, context)
	testMarshal(t, "list",
		"- bar\n- baz",
		[]interface{}{"bar", "baz"},
		context)
	testMarshal(t, "map",
		"foo: bar\nhip: hop",
		map[interface{}]interface{}{"foo": "bar", "hip": "hop"},
		context)
	testMarshal(t, "map with list",
		"foo:\n- bar\n- baz",
		map[interface{}]interface{}{"foo": []interface{}{"bar", "baz"}},
		context)
	testMarshal(t, "map with sub-map",
		"foo:\n  bar:\n    baz",
		map[interface{}]interface{}{"foo": map[interface{}]interface{}{"bar": "baz"}},
		context)
}

func TestMarshalJSONPretty(t *testing.T) {
	context := getvContext{}
	context.asJSON = true
	context.pretty = true

	testMarshal(t, "scalar", `"foo"`, "foo", context)
	testMarshal(t, "scalar number", "1", 1, context)
	testMarshal(t, "list",
		"[\n  \"bar\",\n  \"baz\"\n]",
		[]interface{}{"bar", "baz"},
		context)
	testMarshal(t, "map",
		"{\n  \"foo\": \"bar\",\n  \"hip\": \"hop\"\n}",
		map[interface{}]interface{}{"foo": "bar", "hip": "hop"},
		context)
	testMarshal(t, "map with list",
		"{\n  \"foo\": [\n    \"bar\",\n    \"baz\"\n  ]\n}",
		map[interface{}]interface{}{"foo": []interface{}{"bar", "baz"}},
		context)
	testMarshal(t, "map with sub-map",
		"{\n  \"foo\": {\n    \"bar\": \"baz\"\n  }\n}",
		map[interface{}]interface{}{"foo": map[interface{}]interface{}{"bar": "baz"}},
		context)
}

func TestMarshalKvJSON(t *testing.T) {
	context := getvContext{}
	context.asKvJSON = true
	testMarshal(t, "kvjson map with list",
		"/foo/0: bar\n/foo/1: baz",
		map[interface{}]interface{}{"foo": []interface{}{"bar", "baz"}},
		context)
	testMarshal(t, "kvjson map with sub-map",
		"/foo/bar: baz",
		map[interface{}]interface{}{"foo": map[interface{}]interface{}{"bar": "baz"}},
		context)
}

func TestMarshalPrettyKvJSON(t *testing.T) {
	context := getvContext{}
	context.pretty = true
	context.asKvJSON = true
	testMarshal(t, "kvjson map with list",
		"{\n\"/foo/0\": \"bar\",\n\"/foo/1\": \"baz\"\n}",
		map[interface{}]interface{}{"foo": []interface{}{"bar", "baz"}},
		context)
	testMarshal(t, "kvjson map with sub-map",
		"{\n\"/foo/bar\": \"baz\"\n}",
		map[interface{}]interface{}{"foo": map[interface{}]interface{}{"bar": "baz"}},
		context)
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

	context := getvContext{
		rootContext: &rootContext{
			secretKeyring: *newOptionalString(keyFile, true),
		},
		templateString: *newOptionalString("{{getv \"/username\"}}:{{getv \"/password\"}}", true),
	}
	testMarshal(t, "basic template",
		"foo:bar",
		map[interface{}]interface{}{"username": "foo", "password": "bar"},
		context)

	context.templateString = *newOptionalString("{{cgetv \"/username\"}}:{{cgetv \"/password\"}}", true)
	testMarshal(t, "decrypt template",
		"foo:bar",
		map[interface{}]interface{}{"username": encryptedFoo, "password": encryptedBar},
		context)
}

//
//func TestCsetv(t *testing.T) {
//	tempDir, err := ioutil.TempDir("", "clconf")
//	if err != nil {
//		t.Errorf("Unable to create temp dir: %v", err)
//	}
//	defer func() {
//		os.RemoveAll(tempDir)
//	}()
//
//	file := filepath.Join(tempDir, "config.yml")
//
//	path, expected, actual, err := getCsetvOutcome(file, []string{"/a", "b"}, []string{})
//	if err != nil {
//		t.Errorf("Setv %s encrypt failed: %v", path, err)
//	}
//	if !reflect.DeepEqual(expected, actual) {
//		t.Errorf("Setv %s encrypt failed: %v != %v", path, expected, actual)
//	}
//}
//
//func TestSetv(t *testing.T) {
//	tempDir, err := ioutil.TempDir("", "clconf")
//	if err != nil {
//		t.Errorf("Unable to create temp dir: %v", err)
//	}
//	defer func() {
//		os.RemoveAll(tempDir)
//	}()
//
//	file := filepath.Join(tempDir, "config.yml")
//
//	context := NewTestSetvContext(file, []string{}, []string{})
//	if err := setv(context); err == nil {
//		t.Error("Setv no args should have failed")
//	}
//
//	context = NewTestSetvContext(file, []string{"/a"}, []string{})
//	if err := setv(context); err == nil {
//		t.Error("Setv one arg should have failed")
//	}
//
//	path, expected, actual, err := getSetvOutcome(file, []string{"/a", "b"}, []string{})
//	if err != nil {
//		t.Errorf("Setv %s failed: %v", path, err)
//	}
//	if !reflect.DeepEqual(expected, actual) {
//		t.Errorf("Setv %s failed: %v != %v", path, expected, actual)
//	}
//
//	path, expected, actual, err = getSetvOutcome(file, []string{"/a", "b"}, []string{"--encrypt"})
//	if err != nil {
//		t.Errorf("Setv %s encrypt failed: %v", path, err)
//	}
//	if !reflect.DeepEqual(expected, actual) {
//		t.Errorf("Setv %s encrypt failed: %v != %v", path, expected, actual)
//	}
//}
//
