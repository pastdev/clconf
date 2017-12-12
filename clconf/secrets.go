package clconf

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io/ioutil"

	"github.com/xordataexchange/crypt/encoding/secconf"
)

// SecretAgent loads and holds a keypair needed for
// encryption/decryption
type SecretAgent struct {
	key []byte
}

// Decrypt will return the decrypted value represented by encrypted
func (secretAgent *SecretAgent) Decrypt(encrypted string) (string, error) {
	if secretAgent.key == nil {
        return "", errors.New("SecretAgent missing key")
	}
	b, err := secconf.Decode(
		[]byte(encrypted), 
		bytes.NewBuffer(secretAgent.key))
	if err != nil {
		return "", err
	}
	 
	return string(b), nil
}

// Encrypt will return the encrypted value represented by decrypted
func (secretAgent *SecretAgent) Encrypt(decrypted string) (string, error) {
	if secretAgent.key == nil {
        return "", errors.New("SecretAgent missing key")
	}
	b, err := secconf.Encode(
		[]byte(decrypted), 
		bytes.NewBuffer(secretAgent.key))
	if err != nil {
		return "", err
	}
	 
	return string(b), nil
}

// NewSecretAgent will return a new SecretAgent with the provided
// key.
func NewSecretAgent(key []byte, publicKey []byte) (*SecretAgent) {
    return &SecretAgent{
		key: key,
	}
}

// NewSecretAgentFromFile loads from keyFile
func NewSecretAgentFromFile(keyFile string) (*SecretAgent, error) {
	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	return &SecretAgent{key: key}, nil
}

// NewSecretAgentFromBase64 loads from keyBase64
func NewSecretAgentFromBase64(keyBase64 string) (*SecretAgent, error) {
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return nil, err
	}
	return &SecretAgent{key: key}, nil
}
