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
		Marshaler: Marshaler{
			templateString: *newOptionalString(templateString, true),
		},
	}
	testGetTemplate(t, "template string", "bar", data, context)

	context = getvContext{
		rootContext: &rootContext{
			secretKeyring: *newOptionalString(keyFile, true),
		},
		Marshaler: Marshaler{
			templateBase64: *newOptionalString(templateBase64, true),
		},
	}
	testGetTemplate(t, "template base64", "bar", data, context)

	context = getvContext{
		rootContext: &rootContext{
			secretKeyring: *newOptionalString(keyFile, true),
		},
		Marshaler: Marshaler{
			template: *newOptionalString(templateFile, true),
		},
	}
	testGetTemplate(t, "template file", "bar", data, context)
}
