package cmd

import (
	"errors"
	"os"

	"github.com/pastdev/clconf/v2/clconf"
)

func (c *rootContext) newSecretAgent() (*clconf.SecretAgent, error) {
	var secretAgent *clconf.SecretAgent
	var err error

	if c.secretKeyringBase64.set {
		secretAgent, err = clconf.NewSecretAgentFromBase64(c.secretKeyringBase64.value)
	} else if c.secretKeyring.set {
		secretAgent, err = clconf.NewSecretAgentFromFile(c.secretKeyring.value)
	} else if keyBase64, ok := os.LookupEnv("SECRET_KEYRING_BASE64"); !c.ignoreEnv && ok {
		secretAgent, err = clconf.NewSecretAgentFromBase64(keyBase64)
	} else if keyFile, ok := os.LookupEnv("SECRET_KEYRING"); !c.ignoreEnv && ok {
		secretAgent, err = clconf.NewSecretAgentFromFile(keyFile)
	} else {
		err = errors.New("requires --secret-keyring-base64, --secret-keyring, or SECRET_KEYRING")
	}

	return secretAgent, err
}
