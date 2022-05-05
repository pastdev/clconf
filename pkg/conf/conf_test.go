package conf_test

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/pastdev/clconf/v2/pkg/conf"
	"github.com/pastdev/clconf/v2/pkg/yamljson"
	"github.com/stretchr/testify/assert"
)

func TestBase64Strings(t *testing.T) {
	encoded := []string{}
	actual, err := conf.DecodeBase64Strings(encoded...)
	if err != nil || len(actual) != 0 {
		t.Errorf("Base64Strings empty failed: [%v]", actual)
	}

	expected := []string{"one", "two"}
	encoded = []string{
		base64.StdEncoding.EncodeToString([]byte(expected[0])),
		base64.StdEncoding.EncodeToString([]byte(expected[1]))}
	actual, err = conf.DecodeBase64Strings(encoded...)
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("Base64Strings one two failed: [%v] == [%v]", expected, actual)
	}

	if _, err := conf.DecodeBase64Strings("&*INVALID*&"); err == nil {
		t.Error("Base64Strings invalid should have failed")
	}
}

func TestLoadConf(t *testing.T) {
	envVars := []string{"a"}
	tempDir, err := ioutil.TempDir("", "clconf")
	if err != nil {
		t.Errorf("Unable to create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
		for _, name := range envVars {
			_ = os.Unsetenv(name)
		}
		_ = os.Unsetenv("YAML_FILES")
		_ = os.Unsetenv("YAML_VARS")
		_ = os.Unsetenv("YAML_VAR")
	}()

	envValues := []string{base64.StdEncoding.EncodeToString([]byte("a: env\nenv: env"))}
	for index, name := range envVars {
		assert.Nil(t, os.Setenv(name, envValues[index]))
	}

	stdinFile := path.Join(tempDir, "stdin")
	err = os.WriteFile(stdinFile, []byte("a: stdin\nstdin: 1"), 0600)
	if err != nil {
		t.Errorf("failed to write stdinFile: %v", err)
	}

	fileArg := path.Join(tempDir, "fileArg")
	err = os.WriteFile(fileArg, []byte("a: fileArg\nfileArg: 1"), 0600)
	if err != nil {
		t.Errorf("failed to write fileArg: %v", err)
	}
	fileEnv := path.Join(tempDir, "fileEnv")
	err = os.WriteFile(fileEnv, []byte("a: fileEnv\nfileEnv: 1"), 0600)
	if err != nil {
		t.Errorf("failed to write fileEnv: %v", err)
	}

	b64Arg := base64.StdEncoding.EncodeToString([]byte("a: b64Arg\nb64Arg: 1"))
	b64Env := base64.StdEncoding.EncodeToString([]byte("a: b64Env\nb64Env: 1"))

	actual, err := conf.ConfSources{}.LoadInterface()
	if err != nil || actual == nil {
		t.Errorf("LoadConf no config failed")
	}

	expected, _ := yamljson.UnmarshalYamlInterface("a: b64Arg\nb64Arg: 1")
	actual, err = conf.ConfSources{Overrides: []string{b64Arg}}.LoadInterface()
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf b64Arg failed: [%v] != [%v] (err: %v)", expected, actual, err)
	}

	assert.Nil(t, os.Setenv("YAML_VAR", b64Env))
	assert.Nil(t, os.Setenv("YAML_VARS", "YAML_VAR"))
	expected, _ = yamljson.UnmarshalYamlInterface("a: b64Env\nb64Arg: 1\nb64Env: 1")
	actual, err = conf.ConfSources{Environment: true, Overrides: []string{b64Arg}}.LoadInterface()
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf b64Arg, b64Env failed: [%v] != [%v] (err: %v)", expected, actual, err)
	}
	assert.Nil(t, os.Unsetenv("YAML_VARS"))
	assert.Nil(t, os.Unsetenv("YAML_VAR"))

	assert.Nil(t, os.Setenv("YAML_VAR", b64Env))
	assert.Nil(t, os.Setenv("YAML_VARS", "YAML_VAR"))
	expected, _ = yamljson.UnmarshalYamlInterface("a: b64Arg\nb64Arg: 1")
	actual, err = conf.ConfSources{Overrides: []string{b64Arg}}.LoadInterface()
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf b64Arg, b64Env (but disabled) failed: [%v] != [%v] (err: %v)", expected, actual, err)
	}
	assert.Nil(t, os.Unsetenv("YAML_VARS"))
	assert.Nil(t, os.Unsetenv("YAML_VAR"))

	expected, _ = yamljson.UnmarshalYamlInterface("a: fileArg\nfileArg: 1")
	actual, err = conf.ConfSources{Environment: true, Files: []string{fileArg}}.LoadInterface()
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf fileArg failed: [%v] != [%v] (err: %v)", expected, actual, err)
	}

	assert.Nil(t, os.Setenv("YAML_FILES", fileEnv))
	expected, _ = yamljson.UnmarshalYamlInterface("a: fileEnv\nfileArg: 1\nfileEnv: 1")
	actual, err = conf.ConfSources{Environment: true, Files: []string{fileArg}}.LoadInterface()
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf fileArg, fileEnv failed: [%v] != [%v] (err: %v)", expected, actual, err)
	}
	assert.Nil(t, os.Unsetenv("YAML_FILES"))

	assert.Nil(t, os.Setenv("YAML_VAR", b64Env))
	assert.Nil(t, os.Setenv("YAML_VARS", "YAML_VAR"))
	assert.Nil(t, os.Setenv("YAML_FILES", fileEnv))
	expected, _ = yamljson.UnmarshalYamlInterface("a: b64Env\nfileArg: 1\nfileEnv: 1\nb64Arg: 1\nb64Env: 1")
	actual, err = conf.ConfSources{Environment: true, Files: []string{fileArg}, Overrides: []string{b64Arg}}.LoadInterface()
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf fileArg, fileEnv, b64Arg, b64Env failed: [%v] != [%v] (err: %v)", expected, actual, err)
	}
	assert.Nil(t, os.Unsetenv("YAML_FILES"))
	assert.Nil(t, os.Unsetenv("YAML_VARS"))
	assert.Nil(t, os.Unsetenv("YAML_VAR"))

	stdin, err := os.Open(stdinFile)
	if err != nil {
		t.Errorf("Error opening stdin file for reading")
	}
	defer func() { _ = stdin.Close() }()
	assert.Nil(t, os.Setenv("YAML_VAR", b64Env))
	assert.Nil(t, os.Setenv("YAML_VARS", "YAML_VAR"))
	assert.Nil(t, os.Setenv("YAML_FILES", fileEnv))
	expected, _ = yamljson.UnmarshalYamlInterface("a: stdin\nfileArg: 1\nfileEnv: 1\nb64Arg: 1\nb64Env: 1\nstdin: 1")
	actual, err = conf.ConfSources{
		Files:       []string{fileArg},
		Overrides:   []string{b64Arg},
		Environment: true,
		Stream:      stdin}.LoadInterface()
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf all failed: [%v] != [%v] (err: %v)", expected, actual, err)
	}
	assert.Nil(t, os.Unsetenv("YAML_FILES"))
	assert.Nil(t, os.Unsetenv("YAML_VARS"))
	assert.Nil(t, os.Unsetenv("YAML_VAR"))
}

func TestReadEnvVars(t *testing.T) {
	actual, err := conf.ReadEnvVars()
	if err != nil {
		t.Fatal(err)
	}
	if len(actual) > 0 {
		t.Errorf("ReadEnvVars empty failed")
	}
}

func TestReadEnvVarsDoesNotExist(t *testing.T) {
	_, err := conf.ReadEnvVars("NOT_AN_ENV_VAR_OR_PROBABLY_SHOULDNT_BE")
	if err == nil {
		t.Errorf("ReadEnvVars does not exist should have paniced")
	}
}

func TestReadEnvVarsTempValues(t *testing.T) {
	names := []string{"FOO", "BAZ"}
	values := []string{"bar", "qux"}
	defer func() {
		for _, name := range names {
			_ = os.Unsetenv(name)
		}
	}()

	for index, name := range names {
		assert.Nil(t, os.Setenv(name, values[index]))
	}
	actual, err := conf.ReadEnvVars(names...)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(values, actual) {
		t.Errorf("ReadEnvVars FOO BAZ failed: [%v] [%v]", values, actual)
	}
}

func TestReadFiles(t *testing.T) {
	actual, err := conf.ReadFiles()
	if err != nil || len(actual) > 0 {
		t.Errorf("ReadFiles empty failed")
	}
	if _, err := conf.ReadFiles("NOT_A_FILE_OR_PROBABLY_SHOULDNT_BE"); err == nil {
		t.Errorf("ReadFiles does not exist should have paniced")
	}
}

func TestReadFilesTempValues(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "clconf")
	if err != nil {
		t.Errorf("Unable to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	names := []string{path.Join(tempDir, "foo"), path.Join(tempDir, "baz")}
	values := []string{"bar", "qux"}
	for index, name := range names {
		err := os.WriteFile(name, []byte(values[index]), 0600)
		if err != nil {
			t.Errorf("unable to write %s: %v", name, err)
		}
	}
	actual, err := conf.ReadFiles(names...)
	if err != nil || !reflect.DeepEqual(values, actual) {
		t.Errorf("ReadFiles foo baz failed: [%v] [%v]", values, actual)
	}
}
