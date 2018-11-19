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

	"github.com/pastdev/clconf/clconf"
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

	testGetPath(t, "/", "", *NewOptionalString("", false))
	testGetPath(t, "/", "/", *NewOptionalString("", false))
	testGetPath(t, "/foo", "/foo", *NewOptionalString("", false))

	testGetPath(t, "/foo", "", *NewOptionalString("/foo", true))
	testGetPath(t, "/foo", "/", *NewOptionalString("/foo", true))
	testGetPath(t, "/foo/bar", "/bar", *NewOptionalString("/foo", true))
	testGetPath(t, "/foo/bar", "/bar", *NewOptionalString("/foo/", true))

	os.Setenv(envVar, "/foo")
	testGetPath(t, "/foo", "", *NewOptionalString("", false))
	testGetPath(t, "/foo", "/", *NewOptionalString("", false))
	testGetPath(t, "/foo/bar", "/bar", *NewOptionalString("", false))
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
	context.defaultValue = *NewOptionalString("baz", true)
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
	context.secretKeyring = *NewOptionalString(keyFile, true)
	context.decrypt = []string{"/"}
	testGetValue(t, "decrypt value", "bar", "/foo", context)

	context = testGetvContext(
		fmt.Sprintf("foo: %s\nhip: %s\ntik: tok",
			encryptedBar, encryptedHop))
	context.secretKeyring = *NewOptionalString(keyFile, true)
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
			secretKeyring: *NewOptionalString(keyFile, true),
		},
		templateString: *NewOptionalString(templateString, true),
	}
	testGetTemplate(t, "template string", "bar", data, context)

	context = getvContext{
		rootContext: &rootContext{
			secretKeyring: *NewOptionalString(keyFile, true),
		},
		templateBase64: *NewOptionalString(templateBase64, true),
	}
	testGetTemplate(t, "template base64", "bar", data, context)

	context = getvContext{
		rootContext: &rootContext{
			secretKeyring: *NewOptionalString(keyFile, true),
		},
		template: *NewOptionalString(templateFile, true),
	}
	testGetTemplate(t, "template file", "bar", data, context)
}

//func testGetValueWithTemplate(t *testing.T, name string, args, opts []string) {
//	config, err := NewTestConfig()
//	if err != nil {
//		t.Error(err)
//	}
//
//	context, _, expected, actual, err := getGetvOutcome(config, args, opts)
//
//	_, tmpl, err := getTemplate(context)
//	if err != nil {
//		t.Errorf("GetValueWithTemplate getTemplate %s failed and shouldn't have: %v", name, err)
//	} else if tmpl == nil {
//		t.Errorf("GetValueWithTemplate getTemplate %s expected result", name)
//	}
//
//	expectedString, err := tmpl.Execute(expected)
//	if err != nil {
//		t.Errorf("GetValueWithTemplate tmpl.Exectute %s failed and shouldn't have: %v", name, err)
//	}
//	// when templates are used, the template doesnt get processed
//	// until marshaling...
//	_, actualString, err := marshal(context, actual, nil)
//	if err != nil {
//		t.Errorf("GetValueWithTemplate marshal %s failed and shouldn't have: %v", name, err)
//	}
//
//	if expectedString != actualString {
//		t.Errorf("GetValueWithTemplate %s invalid: [%v] != [%v]",
//			name, expectedString, actualString)
//	}
//}
//
//func TestGetValueWithTemplate(t *testing.T) {
//	tempDir, err := ioutil.TempDir("", "clconf")
//	if err != nil {
//		t.Errorf("Unable to create temp dir: %v", err)
//	}
//	defer func() {
//		os.RemoveAll(tempDir)
//	}()
//
//	templateString := "{{ getv \"/username-plaintext\" }}:{{getv \"/password-plaintext\" }}"
//	templateBytes := []byte(templateString)
//	templateBase64 := base64.StdEncoding.EncodeToString(templateBytes)
//	templateFile := filepath.Join(tempDir, "template")
//	ioutil.WriteFile(templateFile, templateBytes, 0600)
//
//	testGetValueWithTemplate(t, "getv template string",
//		[]string{"/app/db"},
//		[]string{"--template-string", templateString})
//	testGetValueWithTemplate(t, "getv template base64",
//		[]string{"/app/db"},
//		[]string{"--template-base64", templateBase64})
//	testGetValueWithTemplate(t, "getv template file",
//		[]string{"/app/db"},
//		[]string{"--template", templateFile})
//
//	testGetValueWithTemplate(t, "cgetv template string",
//		[]string{"/app/db"},
//		[]string{
//			"--template-string", "{{ cgetv \"/username\" }}:{{cgetv \"/password\" }}",
//		})
//}
//
//func TestMarshal(t *testing.T) {
//	var expected interface{}
//	var actual interface{}
//
//	context := NewTestGlobalContext()
//	expected = "foo"
//	context, actual, err := marshal(context, expected, nil)
//	if context != context || actual != expected || err != nil {
//		t.Errorf("Marshal string failed: [%v] [%v != %v] [%v]", context, actual, expected, err)
//	}
//
//	expected = "2"
//	context, actual, err = marshal(context, expected, nil)
//	if context != context || actual != expected || err != nil {
//		t.Errorf("Marshal int failed: [%v] [%v != %v] [%v]", context, actual, expected, err)
//	}
//
//	expected, _ = UnmarshalYaml("a:\n  b: foo")
//	context, marshaled, err := marshal(context, expected, nil)
//	actual, _ = UnmarshalYaml(marshaled)
//	if context != context || !reflect.DeepEqual(actual, expected) || err != nil {
//		t.Errorf("Marshal map failed: [%v] [%v != %v] [%v]", context, actual, expected, err)
//	}
//
//	expected, _ = UnmarshalYaml("a:\n- foo\n- bar")
//	context, marshaled, err = marshal(context, expected, nil)
//	actual, _ = UnmarshalYaml(marshaled)
//	if context != context || !reflect.DeepEqual(actual, expected) || err != nil {
//		t.Errorf("Marshal array failed: [%v] [%v != %v] [%v]", context, actual, expected, err)
//	}
//}
//
//func TestNewSecretAgentFromCli(t *testing.T) {
//	var err error
//	secretKeyringEnvVar := "SECRET_KEYRING"
//	secretKeyringBase64EnvVar := "SECRET_KEYRING_BASE64"
//	defer func() {
//		// just in case
//		os.Unsetenv(secretKeyringEnvVar)
//		os.Unsetenv(secretKeyringBase64EnvVar)
//	}()
//
//	_, err = newSecretAgentFromCli(
//		NewTestContext(Name, nil, globalFlags(), nil, []string{}, []string{}))
//	if err == nil {
//		t.Errorf("New secret agent no options no env failed: [%v]", err)
//	}
//
//	secretAgent, err := newSecretAgentFromCli(
//		NewTestContext(Name, nil, globalFlags(), nil, []string{},
//			[]string{"--secret-keyring", NewTestKeysFile()}))
//	if err != nil || secretAgent.key == nil {
//		t.Errorf("New secret agent from file failed: [%v]", err)
//	}
//
//	secretKeyring, err := ioutil.ReadFile(NewTestKeysFile())
//	if err != nil {
//		t.Errorf("New secret agent from base 64 read keys file failed: [%v]", err)
//	}
//	secretAgent, err = newSecretAgentFromCli(
//		NewTestContext(Name, nil, globalFlags(), nil, []string{},
//			[]string{"--secret-keyring-base64", base64.StdEncoding.EncodeToString(secretKeyring)}))
//	if err != nil || secretAgent.key == nil {
//		t.Errorf("New secret agent from base 64 failed: [%v]", err)
//	}
//
//	err = os.Setenv(secretKeyringEnvVar, NewTestKeysFile())
//	if err != nil {
//		t.Errorf("New secret agent from env set env failed: [%v]", err)
//	}
//	secretAgent, err = newSecretAgentFromCli(
//		NewTestContext(Name, nil, globalFlags(), nil, []string{}, []string{}))
//	if err != nil || secretAgent.key == nil {
//		t.Errorf("New secret agent from env failed: [%v]", err)
//	}
//	os.Unsetenv(secretKeyringEnvVar)
//
//	err = os.Setenv(secretKeyringBase64EnvVar,
//		base64.StdEncoding.EncodeToString(secretKeyring))
//	if err != nil {
//		t.Errorf("New secret agent from base 64 env set env failed: [%v]", err)
//	}
//	secretAgent, err = newSecretAgentFromCli(
//		NewTestContext(Name, nil, globalFlags(), nil, []string{}, []string{}))
//	if err != nil || secretAgent.key == nil {
//		t.Errorf("New secret agent from base 64 env failed: [%v]", err)
//	}
//	os.Unsetenv(secretKeyringEnvVar)
//}
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
