package clconf

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"

	"github.com/xordataexchange/crypt/encoding/secconf"
)

// SecretAgent loads and holds a keypair needed for
// encryption/decryption
type SecretAgent struct {
	privateKey []byte
	publicKey []byte
}

// Decrypt will return the decrypted value represented by encrypted
func (secretAgent *SecretAgent) Decrypt(encrypted string) (string, error) {
	b, err := secconf.Decode(
		[]byte(encrypted), 
		bytes.NewBuffer(secretAgent.privateKey))
	if err != nil {
		return "", err
	}
	 
	return string(b), nil
}

// Encrypt will return the encrypted value represented by decrypted
func (secretAgent *SecretAgent) Encrypt(decrypted string) (string, error) {
	b, err := secconf.Encode(
		[]byte(decrypted), 
		bytes.NewBuffer(secretAgent.privateKey))
	if err != nil {
		return "", err
	}
	 
	return string(b), nil
}

func keyBytes(path string, conf interface{}) ([]byte, bool) {
	key, ok := GetValue(path, conf)
	if !ok {
        return nil, false
	}
	keyString, ok := key.(string);
	if !ok {
        return nil, false
	}
	return []byte(keyString), true
}

func newSecretAgent(yaml []byte, err error) (*SecretAgent, error) {
	if err != nil {
		return nil, err
	}

	config, err := unmarshalYaml(yaml)
	if err != nil  {
		return nil, err
	}

	privateKey, ok := keyBytes("/private-key", config)
	if !ok {
        return nil, err
	}
	publicKey, ok := keyBytes("/public-key", config)
	if !ok {
        return nil, err
	}

    return NewSecretAgent(privateKey, publicKey), nil
}

// NewSecretAgent will return a new SecretAgent with the provided
// privateKey.
func NewSecretAgent(privateKey []byte, publicKey []byte) (*SecretAgent) {
    return &SecretAgent{
		privateKey: privateKey,
		publicKey: publicKey,
	}
}

// NewSecretAgentFromBase64 will load the private key from the supplied
// base64 encoded string.
func NewSecretAgentFromBase64(base64String string) (*SecretAgent, error) {
	return newSecretAgent(base64.StdEncoding.DecodeString(base64String))
}

// NewSecretAgentFromFile will load the private key from the file indicated
// by path.
func NewSecretAgentFromFile(path string) (*SecretAgent, error) {
	return newSecretAgent(ioutil.ReadFile(path))
}
