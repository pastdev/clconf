package clconf

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/urfave/cli"
)

const (
	// Name is the name of this application
	Name = "clconf"
	// Version is the version of this application
	Version = "0.0.1"
)

// https://stackoverflow.com/a/26804949/516433
var print = fmt.Print

func cgetv(c *cli.Context) error {
	return dump(marshal(cgetvHandler(c)))
}

func cgetvFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{Name: "default"},
	}
}

func cgetvHandler(c *cli.Context) (*cli.Context, interface{}, cli.ExitCoder) {
	_, value, err := getValue(c)
	if err != nil {
		return c, nil, err
	}
	secretAgent, err := newSecretAgentFromCli(c)
	if err != nil {
		return c, nil, err
	}
	if stringValue, ok := value.(string); ok {
		decrypted, err := secretAgent.Decrypt(stringValue)
		return c, decrypted, cliError(err, 1)
	}
	return c, nil, cli.NewExitError("value at specified path not a string", 1)
}

func cliError(err error, exitCode int) cli.ExitCoder {
	if err != nil {
		if casted, ok := err.(cli.ExitCoder); ok {
			return casted
		}
		return cli.NewExitError(err, exitCode)
	}
	return nil
}

func csetv(c *cli.Context) error {
	if err := c.Set("encrypt", "true"); err != nil {
		return cli.NewExitError(err, 1)
	}
	return setv(c)
}

func dump(c *cli.Context, value interface{}, err cli.ExitCoder) cli.ExitCoder {
	if err != nil {
		return err
	}
	print(value)
	return nil
}

func getDefault(c *cli.Context) (string, bool) {
	if defaultValue := c.String("default"); defaultValue != "" {
		return defaultValue, true
	}
	return "", false
}

func getPath(c *cli.Context) string {
	valuePath := c.Args().First()

	if prefix := c.GlobalString("prefix"); prefix != "" {
		return path.Join(prefix, valuePath)
	} else if prefix, ok := os.LookupEnv("CONFIG_PREFIX"); ok {
		return path.Join(prefix, valuePath)
	}

	if valuePath == "" {
		return "/"
	}
	return valuePath
}

func getv(c *cli.Context) error {
	return dump(marshal(getValue(c)))
}

func getValue(c *cli.Context) (*cli.Context, interface{}, cli.ExitCoder) {
	path := getPath(c)
	config, err := load(c)
	if err != nil {
		return c, nil, cliError(err, 1)
	}
	value, ok := GetValue(config, path)
	if !ok {
		value, ok = getDefault(c)
		if !ok {
			return c, nil, cli.NewExitError(fmt.Sprintf("[%v] does not exist", path), 1)
		}
	}
	if decryptPaths := c.StringSlice("decrypt"); len(decryptPaths) > 0 {
		secretAgent, err := newSecretAgentFromCli(c)
		if err != nil {
			return c, nil, err
		}
		err = cliError(secretAgent.DecryptPaths(value, decryptPaths...), 1)
		if err != nil {
			return c, nil, err
		}
	}
	return c, value, nil
}

func getvFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringSliceFlag{Name: "decrypt"},
		cli.StringFlag{Name: "default"},
	}
}

func globalFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{Name: "prefix"},
		cli.StringFlag{Name: "public-keyring"},
		cli.StringFlag{Name: "public-keyring-base64"},
		cli.StringFlag{Name: "secret-keyring"},
		cli.StringFlag{Name: "secret-keyring-base64"},
		cli.StringSliceFlag{Name: "yaml"},
		cli.StringSliceFlag{Name: "yaml-base64"},
	}
}

func load(c *cli.Context) (map[interface{}]interface{}, cli.ExitCoder) {
	config, err := LoadConfFromEnvironment(
		c.GlobalStringSlice("yaml"),
		c.GlobalStringSlice("yaml-base64"))
	return config, cliError(err, 1)
}

func loadForSetv(c *cli.Context) (string, map[interface{}]interface{}, cli.ExitCoder) {
	path, config, err := LoadSettableConfFromEnvironment(c.GlobalStringSlice("yaml"))
	return path, config, cliError(err, 1)
}

func marshal(c *cli.Context, value interface{}, err cli.ExitCoder) (*cli.Context, string, cli.ExitCoder) {
	if err != nil {
		return c, "", err
	}
	if stringValue, ok := value.(string); ok {
		return c, stringValue, nil
	} else if mapValue, ok := value.(map[interface{}]interface{}); ok {
		marshaled, err := MarshalYaml(mapValue)
		return c, string(marshaled), cliError(err, 1)
	} else if arrayValue, ok := value.([]interface{}); ok {
		marshaled, err := MarshalYaml(arrayValue)
		return c, string(marshaled), cliError(err, 1)
	}
	return c, fmt.Sprintf("%v", value), err
}

// NewApp returns a new cli application instance ready to be run.
//
// Thoughts...
// ** stdin should be read in as a file and used as override (not sure if before or after --overrides)
func NewApp() *cli.App {
	app := cli.NewApp()
	app.Name = Name
	app.Version = Version
	app.UsageText = "clconf [global options] command [command options] [args...]"

	app.Flags = globalFlags()

	app.Commands = []cli.Command{
		{
			Name:      "cgetv",
			Usage:     "Get a secret value",
			ArgsUsage: "PATH",
			Action:    cgetv,
			Flags:     cgetvFlags(),
		},
		{
			Name:      "getv",
			Usage:     "Get a value",
			ArgsUsage: "PATH",
			Action:    getv,
			Flags:     getvFlags(),
		},
		{
			Name:      "csetv",
			Usage:     "Set a secret value",
			ArgsUsage: "PATH VALUE",
			Action:    csetv,
			Flags:     setvFlags(),
		},
		{
			Name:      "setv",
			Usage:     "Set a value",
			ArgsUsage: "PATH VALUE",
			Action:    setv,
			Flags:     setvFlags(),
		},
	}

	app.Action = getv

	return app
}

func newSecretAgentFromCli(c *cli.Context) (*SecretAgent, cli.ExitCoder) {
	var err error
	var secretAgent *SecretAgent

	if keyBase64 := c.GlobalString("secret-keyring-base64"); keyBase64 != "" {
		secretAgent, err = NewSecretAgentFromBase64(keyBase64)
	} else if keyFile := c.GlobalString("secret-keyring"); keyFile != "" {
		secretAgent, err = NewSecretAgentFromFile(keyFile)
	} else if keyBase64, ok := os.LookupEnv("SECRET_KEYRING_BASE64"); ok {
		secretAgent, err = NewSecretAgentFromBase64(keyBase64)
	} else if keyFile, ok := os.LookupEnv("SECRET_KEYRING"); ok {
		secretAgent, err = NewSecretAgentFromFile(keyFile)
	} else {
		err = errors.New("requires --secret-keyring-base64, --secret-keyring, or SECRET_KEYRING")
	}

	return secretAgent, cliError(err, 1)
}

func setv(c *cli.Context) error {
	if c.NArg() != 2 {
		return cli.NewExitError("setv requires path and value args", 1)
	}

	path := getPath(c)
	value := c.Args().Get(1)
	file, config, err := loadForSetv(c)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to load config: %v", err), 1)
	}

	if c.Bool("encrypt") {
		secretAgent, err := newSecretAgentFromCli(c)
		if err != nil {
			return err
		}
		encrypted, encryptErr := secretAgent.Encrypt(value)
		if encryptErr != nil {
			return cli.NewExitError(
				fmt.Sprintf("Failed to encrypt value: %v", err), 1)
		}
		value = encrypted
	}

	if err := SetValue(config, path, value); err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to load config: %v", err), 1)
	}

	return cliError(SaveConf(config, file), 1)
}

func setvFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{Name: "encrypt"},
	}
}
