package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/pastdev/clconf/clconf"
	yaml "gopkg.in/yaml.v2"
)

func getSetValueActual(message, original, key, value string, context *setvContext) (string, error) {
	file := context.yaml[0]
	err := ioutil.WriteFile(file, []byte(original), 0700)
	if err != nil {
		return "", fmt.Errorf("%s failed to write original to [%s]: %s",
			message, file, err)
	}

	err = context.setValue(key, value)
	if err != nil {
		return "", fmt.Errorf("%s failed to set value [%s] to [%s]: %s",
			message, key, value, err)
	}

	actual, err := ioutil.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("%s failed to read yaml [%s]: %s",
			message, file, err)
	}

	return string(actual), nil
}

func testSetValue(t *testing.T, message, expected, original, key, value string, context *setvContext) {
	actual, err := getSetValueActual(
		fmt.Sprintf("testSetValue %s", message), original, key, value, context)
	if err != nil {
		t.Error(err)
	}

	assertYamlEqual(t, fmt.Sprintf("testSetValue %s", message), expected, actual)
}

func TestSetValue(t *testing.T) {
	context := &setvContext{rootContext: &rootContext{}}
	if err := context.setValue("/foo", "bar"); err == nil {
		t.Error("setValue no yaml should have failed")
	}

	tempDir, err := ioutil.TempDir("", "clconf")
	if err != nil {
		t.Errorf("Unable to create temp dir: %v", err)
	}
	defer func() {
		os.RemoveAll(tempDir)
	}()
	file := filepath.Join(tempDir, "config.yml")

	keyFile := path.Join("..", "testdata", "test.secring.gpg")
	rootContext := &rootContext{
		secretKeyring: *newOptionalString(keyFile, true),
		yaml:          []string{file},
	}

	context = &setvContext{rootContext: rootContext}
	testSetValue(t, "empty yaml", "foo: bar", "", "/foo", "bar", context)
	testSetValue(t, "new value",
		"foo: bar\nhip: hop",
		"hip: hop",
		"/foo", "bar",
		context)
	testSetValue(t, "replace value",
		"foo: baz",
		"foo: bar",
		"/foo", "baz",
		context)
	testSetValue(t, "new sub value",
		"foo:\n  bar: baz\n  hip: hop",
		"foo:\n  bar: baz",
		"/foo/hip", "hop",
		context)
	testSetValue(t, "new sub value without existing parent",
		"foo:\n  bar:\n    hip: hop",
		"foo:",
		"/foo/bar/hip", "hop",
		context)

	secretAgent, err := clconf.NewSecretAgentFromFile(keyFile)
	if err != nil {
		t.Errorf("Unable to load secret agent %s: %s", keyFile, err)
	}
	context = &setvContext{
		rootContext: rootContext,
		encrypt:     true,
	}
	expected := "bar"
	actualYaml, err := getSetValueActual("encrypted", "", "/foo", expected, context)
	if err != nil {
		t.Error("Unable to encrypt bar")
	}
	var unmarshaled map[string]string
	err = yaml.Unmarshal([]byte(actualYaml), &unmarshaled)
	if err != nil {
		t.Errorf("Unable to unmarshal yaml %s: %s", err, actualYaml)
		return
	}
	actual, err := secretAgent.Decrypt(unmarshaled["foo"])
	if expected != actual {
		t.Errorf("SetValue encrypted not equal [%s] != [%s]", expected, actual)
	}
}
