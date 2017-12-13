package clconf

import (
	"os"
	"encoding/base64"
	"flag"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/urfave/cli"
)

func NewTestConfig() (interface{}, error) {
	config, err := NewTestConfigContent()
	if err != nil {
		return "", err
	}
	return unmarshalYaml(config)
}

func NewTestConfigFile() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "testconfig.yml")
}

func NewTestConfigContent() ([]byte, error) {
	return ioutil.ReadFile(NewTestConfigFile())
}

func NewTestConfigBase64() (string, error) {
	config, err := NewTestConfigContent()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString([]byte(config)), nil
}

func NewTestContext(name string, app *cli.App, flags []cli.Flag, parentContext *cli.Context, args ...string) *cli.Context {
	set := flag.NewFlagSet(name, 0)
	for _, flag := range globalFlags() {
		flag.Apply(set)
	}
	context := cli.NewContext(app, set, parentContext)
	set.Parse(args)
	return context
}

func NewTestGlobalContext() *cli.Context {
	context := NewTestContext(Name, nil, globalFlags(), nil,
		"--secret-keyring", NewTestKeysFile(),
		"--yaml", NewTestConfigFile(),
	)
	return context
}

func NewGetvContext(args ...string) *cli.Context {
	context := NewTestContext("getv", nil, getvFlags(), NewTestGlobalContext(), args...)
	return context
}

func testCgetvHandler(t *testing.T, config interface{}, path string) {
	expected, ok := GetValue(path+"-plaintext", config)

	_, actual, err := cgetvHandler(NewGetvContext(path))
	if ok && err != nil {
		t.Errorf("Cgetv %s failed and shouldn't have: %v", path, err)
	} else if !ok && err == nil {
		t.Errorf("Cgetv %s didn't fail and should have", path)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Cgetv %s unexpected result: %v != %v", path, expected, actual)
	}
}

func TestCgetvHandler(t *testing.T) {
	config, err := NewTestConfig()
	if err != nil {
		t.Error(err)
	}

	testCgetvHandler(t, config, "")
	testCgetvHandler(t, config, "/app/db/username")
	testCgetvHandler(t, config, "/app/db/password")
	testCgetvHandler(t, config, "INVALID_PATH")
}

func testGetPath(t *testing.T, expected string, args ...string) {
	path := getPath(NewTestContext("test", nil, globalFlags(), nil, args...))
	if expected != path {
		t.Errorf("Get path failed: [%v] != [%v]", expected, path)
	}
}

func TestGetPath(t *testing.T) {
	var envVar = "CONFIG_PREFIX"
	defer func() {
        os.Unsetenv(envVar)
	}()

	testGetPath(t, "/", "")
	testGetPath(t, "/", "/")
	testGetPath(t, "/foo", "/foo")

	testGetPath(t, "/foo", "--prefix", "/foo")
	testGetPath(t, "/foo", "--prefix", "/foo", "/")
	testGetPath(t, "/foo/bar", "--prefix", "/foo", "/bar")
	testGetPath(t, "/foo/bar", "--prefix", "/foo/", "/bar")

	os.Setenv(envVar, "/foo")
	testGetPath(t, "/foo", "")
	testGetPath(t, "/foo", "/")
	testGetPath(t, "/foo/bar", "/bar")
}

func testGetValue(t *testing.T, config interface{}, args ...string) {
	context := NewGetvContext(args...)
	path := getPath(context)
	expected, ok := GetValue(path, config)
	if !ok {
		if defaultValue, defaultOk := getDefault(context); defaultOk {
			expected = defaultValue
			ok = true
		}
	}

	_, actual, err := getValue(context)
	if ok && err != nil {
		t.Errorf("Getv %s failed and shouldn't have: %v", path, err)
	} else if !ok && err == nil {
		t.Errorf("Getv %s didn't fail and should have", path)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Getv %s unexpected result: %v != %v", path, expected, actual)
	}
}

func TestGetValue(t *testing.T) {
	config, err := NewTestConfig()
	if err != nil {
		t.Error(err)
	}

	testGetValue(t, config, "")
	testGetValue(t, config, "/")
	testGetValue(t, config, "/app")
	testGetValue(t, config, "/app/db")
	testGetValue(t, config, "/app/db/hostname")
	testGetValue(t, config, "/app/db/hostname", "--default", "INVALID_HOST")
	testGetValue(t, config, "INVALID_PATH")
	testGetValue(t, config, "INVALID_PATH_WITH_DEFAULT", "--default", "foo")
}

func TestMarshal(t *testing.T) {
	var expected interface{}
	var actual interface{}

	expected = "foo"
	context, actual, err := marshal(nil, expected, nil)
	if context != nil || actual != expected || err != nil {
		t.Errorf("Marshal string failed: [%v] [%v != %v] [%v]", context, actual, expected, err)
	}

	expected = "2"
	context, actual, err = marshal(nil, expected, nil)
	if context != nil || actual != expected || err != nil {
		t.Errorf("Marshal int failed: [%v] [%v != %v] [%v]", context, actual, expected, err)
	}

	expected, _ = UnmarshalYaml("a:\n  b: foo")
	context, marshaled, err := marshal(nil, expected, nil)
	actual, _ = UnmarshalYaml(marshaled)
	if context != nil || !reflect.DeepEqual(actual, expected) || err != nil {
		t.Errorf("Marshal map failed: [%v] [%v != %v] [%v]", context, actual, expected, err)
	}

	expected, _ = UnmarshalYaml("a:\n- foo\n- bar")
	context, marshaled, err = marshal(nil, expected, nil)
	actual, _ = UnmarshalYaml(marshaled)
	if context != nil || !reflect.DeepEqual(actual, expected) || err != nil {
		t.Errorf("Marshal array failed: [%v] [%v != %v] [%v]", context, actual, expected, err)
	}
}

func TestNewSecretAgentFromCli(t *testing.T) {
	var err error;
	secretKeyringEnvVar := "SECRET_KEYRING"
	secretKeyringBase64EnvVar := "SECRET_KEYRING_BASE64"
	defer func() {
		// just in case
        os.Unsetenv(secretKeyringEnvVar);
        os.Unsetenv(secretKeyringBase64EnvVar);
	}()

	_, err = newSecretAgentFromCli(
		NewTestContext(Name, nil, globalFlags(), nil))
	if err == nil {
		t.Errorf("New secret agent no options no env failed: [%v]", err)
	} 

	secretAgent, err := newSecretAgentFromCli(
		NewTestContext(Name, nil, globalFlags(), nil,
			"--secret-keyring", NewTestKeysFile()))
	if err != nil || secretAgent.key == nil {
		t.Errorf("New secret agent from file failed: [%v]", err)
	}

	secretKeyring, err := ioutil.ReadFile(NewTestKeysFile())
	if err != nil {
		t.Errorf("New secret agent from base 64 read keys file failed: [%v]", err)
	}
	secretAgent, err = newSecretAgentFromCli(
		NewTestContext(Name, nil, globalFlags(), nil,
			"--secret-keyring-base64", base64.StdEncoding.EncodeToString(secretKeyring)))
	if err != nil || secretAgent.key == nil {
		t.Errorf("New secret agent from base 64 failed: [%v]", err)
	}

	err = os.Setenv(secretKeyringEnvVar, NewTestKeysFile())
	if err != nil {
		t.Errorf("New secret agent from env set env failed: [%v]", err)
	}
	secretAgent, err = newSecretAgentFromCli(
		NewTestContext(Name, nil, globalFlags(), nil))
	if err != nil || secretAgent.key == nil {
		t.Errorf("New secret agent from env failed: [%v]", err)
	}
	os.Unsetenv(secretKeyringEnvVar)

	err = os.Setenv(secretKeyringBase64EnvVar,
		base64.StdEncoding.EncodeToString(secretKeyring))
	if err != nil {
		t.Errorf("New secret agent from base 64 env set env failed: [%v]", err)
	}
	secretAgent, err = newSecretAgentFromCli(
		NewTestContext(Name, nil, globalFlags(), nil))
	if err != nil || secretAgent.key == nil {
		t.Errorf("New secret agent from base 64 env failed: [%v]", err)
	}
	os.Unsetenv(secretKeyringEnvVar)
}
