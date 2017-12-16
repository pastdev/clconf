package clconf

import (
	"encoding/base64"
	"io/ioutil"
	"path/filepath"
	"runtime"
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

func NewTestKeysFile() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "test.secring.gpg")
}

func NewTestSecretAgent() (*SecretAgent, error) {
	return NewSecretAgentFromFile(NewTestKeysFile())
}

func ValuesAtPathsAreEqual(config interface{}, a, b string) bool {
	aValue, ok := GetValue(config, a)
	if !ok {
		return false
	}
	bValue, ok := GetValue(config, b)
	if !ok {
		return false
	}
	return aValue == bValue
}
