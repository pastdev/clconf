package secret

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pastdev/clconf/v2/pkg/core"
	"github.com/pastdev/clconf/v2/pkg/yamljson"
)

func NewTestConfig() (interface{}, error) {
	config, err := NewTestConfigContent()
	if err != nil {
		return "", err
	}
	v, err := yamljson.UnmarshalYamlInterface(string(config))
	if err != nil {
		return "", fmt.Errorf("unmarshal: %w", err)
	}
	return v, nil
}

func NewTestConfigContent() ([]byte, error) {
	v, err := os.ReadFile(NewTestConfigFile())
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	return v, nil
}

func NewTestConfigFile() string {
	return filepath.Join("..", "..", "testdata", "testconfig.yml")
}

func NewTestKeysFile() string {
	return filepath.Join("..", "..", "testdata", "test.secring.gpg")
}

func NewTestSecretAgent() (*SecretAgent, error) {
	return NewSecretAgentFromFile(NewTestKeysFile())
}

func ValuesAtPathsAreEqual(config interface{}, a, b string) bool {
	aValue, err := core.GetValue(config, a)
	if err != nil {
		return false
	}
	bValue, err := core.GetValue(config, b)
	if err != nil {
		return false
	}
	return aValue == bValue
}
