package clconf

import (
	"path/filepath"
	"runtime"
	"testing"
)

func NewTestKeysFile() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "testkeys.yml")
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
