package clconf

import (
	"io/ioutil"
	"path/filepath"
)

func NewTestConfig() (interface{}, error) {
	config, err := NewTestConfigContent()
	if err != nil {
		return "", err
	}
	return unmarshalYaml(config)
}

func NewTestConfigContent() ([]byte, error) {
	return ioutil.ReadFile(NewTestConfigFile())
}

func NewTestConfigFile() string {
	return filepath.Join("..", "testdata", "testconfig.yml")
}

func NewTestKeysFile() string {
	return filepath.Join("..", "testdata", "test.secring.gpg")
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
