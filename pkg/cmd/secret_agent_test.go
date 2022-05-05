package cmd

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/pastdev/clconf/v3/pkg/secret"
	"github.com/stretchr/testify/assert"
)

func testNewSecretAgent(
	t *testing.T,
	message string,
	expected string, //nolint:unparam
	encrypted string,
	context *rootContext,
) {
	secretAgent, err := context.newSecretAgent()
	if err != nil {
		t.Errorf("testNewSecretAgent %s unable to create: %s", message, err)
	}

	actual, err := secretAgent.Decrypt(encrypted)
	if err != nil {
		t.Errorf("testNewSecretAgent %s unable to decrypt: %s", message, err)
	}
	if expected != actual {
		t.Errorf("testNewSecretAgent %s: %s != %s", message, expected, actual)
	}
}

func TestNewSecretAgent(t *testing.T) {
	var err error
	secretKeyringEnvVar := "SECRET_KEYRING"
	secretKeyringBase64EnvVar := "SECRET_KEYRING_BASE64"
	defer func() {
		_ = os.Unsetenv(secretKeyringEnvVar)
		_ = os.Unsetenv(secretKeyringBase64EnvVar)
	}()

	keyFile := path.Join("..", "..", "testdata", "test.secring.gpg")
	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		t.Errorf("Unable to read %s: %s", keyFile, err)
	}
	keyBase64 := base64.StdEncoding.EncodeToString([]byte(key))
	secretAgent := secret.NewSecretAgent(key)
	expected := "foo"
	encrypted, err := secretAgent.Encrypt(expected)
	if err != nil {
		t.Errorf("Unable to encrypt %s", expected)
	}

	_, err = (&rootContext{}).newSecretAgent()
	if err == nil {
		t.Error("newSecretAgent with no key should fail")
	}

	testNewSecretAgent(t, "file", expected, encrypted,
		&rootContext{
			secretKeyring: *newOptionalString(keyFile, true),
		})
	testNewSecretAgent(t, "base64", expected, encrypted,
		&rootContext{
			secretKeyringBase64: *newOptionalString(keyBase64, true),
		})

	err = os.Setenv(secretKeyringEnvVar, keyFile)
	if err != nil {
		t.Errorf("Unable to set env var %s: %s", secretKeyringEnvVar, err)
	}
	testNewSecretAgent(t, "file env var", expected, encrypted, &rootContext{})
	assert.Nil(t, os.Unsetenv(secretKeyringEnvVar))

	err = os.Setenv(secretKeyringBase64EnvVar, keyBase64)
	if err != nil {
		t.Errorf("Unable to set env var %s: %s", secretKeyringBase64EnvVar, err)
	}
	testNewSecretAgent(t, "base64 env var", expected, encrypted, &rootContext{})
	assert.Nil(t, os.Unsetenv(secretKeyringEnvVar))
}
