package clconf_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"gitlab.com/pastdev/s2i/clconf/clconf"
)

func testSecretKeysFile() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "testkeys.yml")
}

func testSecretAgent() (*clconf.SecretAgent, error) {
	return clconf.NewSecretAgentFromFile(testSecretKeysFile())
}

func TestEncryptDecrypt(t *testing.T) {
	plaintext := "foobar"
	secretAgent, err := testSecretAgent()
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
