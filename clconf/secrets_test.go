package clconf

import (
	"encoding/base64"
	"reflect"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"testing"
)

func NewTestKeysFile() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "test.secring.gpg")
}

func NewTestSecretAgent() (*SecretAgent, error) {
	return NewSecretAgentFromFile(NewTestKeysFile())
}

func TestEncryptDecrypt(t *testing.T) {
	plaintext := "SECRET"
	secretAgent, err := NewTestSecretAgent()
	if err != nil {
		t.Errorf("Unable to create secret agent: %v", err)
	}
	ciphertext, err := secretAgent.Encrypt(plaintext)
	if err != nil {
		t.Errorf("Unable to encrypt: %v", err)
	}
	decrypted, err := secretAgent.Decrypt(ciphertext)
	if err != nil {
		t.Errorf("Unable to decrypt: %v", err)
	}
	if decrypted != plaintext {
		t.Errorf("Decrypted doesnt match plaintext: %v", decrypted)
	}
}

func TestNewSecretAgent(t *testing.T) {
	expected, err := ioutil.ReadFile(NewTestKeysFile())
	if err != nil {
		t.Errorf("Unable to read key file: %v", err)
	}

	secretAgent, err := NewSecretAgentFromFile(NewTestKeysFile())
	if err != nil {
		t.Errorf("Unable to create secret agent from file: %v", err)
	}
	if !reflect.DeepEqual(expected, secretAgent.key) {
		t.Errorf("Unable to create secret agent from file: %v", err)
	}

	secretAgent, err = NewSecretAgentFromBase64(base64.StdEncoding.EncodeToString(expected))
	if err != nil {
		t.Errorf("Unable to create secret agent from base64: %v", err)
	}
	if !reflect.DeepEqual(expected, secretAgent.key) {
		t.Errorf("Unable to create secret agent from base64: %v", err)
	}
}
