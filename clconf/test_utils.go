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
	return UnmarshalYaml(string(config))
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
	aValue, err := GetValue(config, a)
	if err != nil {
		return false
	}
	bValue, err := GetValue(config, b)
	if err != nil {
		return false
	}
	return aValue == bValue
}
