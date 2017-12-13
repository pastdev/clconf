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
	return cli.NewExitError("Not yet implemented", 1)
}

func dump(c *cli.Context, value interface{}, err cli.ExitCoder) cli.ExitCoder {
	if err != nil {
		return err
	}
	print(value)
	return nil
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
	value, ok := GetValue(path, config)
	if !ok {
		return c, nil, cli.NewExitError(fmt.Sprintf("[%v] does not exist", path), 1)
	}
	return c, value, nil
}

func getvFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{Name: "default"},
	}
}

func globalFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{Name: "prefix"},
		cli.StringFlag{Name: "public-keyring"},
		cli.StringFlag{Name: "public-keyring-b64"},
		cli.StringFlag{Name: "secret-keyring"},
		cli.StringFlag{Name: "secret-keyring-b64"},
		cli.StringSliceFlag{Name: "yaml"},
		cli.StringSliceFlag{Name: "yaml-b64"},
	}
}

func load(c *cli.Context) (map[interface{}]interface{}, cli.ExitCoder) {
	config, err := LoadConfFromEnvironment(
		c.GlobalStringSlice("yaml"),
		c.GlobalStringSlice("yaml-b64"))
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
// * there should be actions `getv`, `cgetv`, `setv`, `csetv`
// ** all take a coordinate, setters also take a _fileish_ thing
// *** setter _fileish_ thing could be cached as env var...
// ** `getv` should allow for a --decrypt option which takes a list of coords to decrypt
// * --override opition (multi valued) takes base64 encoded yaml
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
			Flags: getvFlags(),
		},
		{
			Name:      "getv",
			Usage:     "Get a value",
			ArgsUsage: "PATH",
			Action:    getv,
			Flags: getvFlags(),
		},
		{
			Name:      "csetv",
			Usage:     "Set a secret value",
			ArgsUsage: "PATH",
			Action:    csetv,
		},
		{
			Name:      "setv",
			Usage:     "Set a value",
			ArgsUsage: "PATH",
			Action:    setv,
		},
	}

	app.Action = getv

	return app
}

func newSecretAgentFromCli(c *cli.Context) (*SecretAgent, cli.ExitCoder) {
	var err error
	var secretAgent *SecretAgent

	if keyBase64 := c.GlobalString("secret-keyring-b64"); keyBase64 != "" {
		secretAgent, err = NewSecretAgentFromBase64(keyBase64)
	} else if keyFile := c.GlobalString("secret-keyring"); keyFile != "" {
		secretAgent, err = NewSecretAgentFromFile(keyFile)
	} else if keyFile, ok := os.LookupEnv("SECRET_KEYRING"); ok {
		secretAgent, err = NewSecretAgentFromFile(keyFile)
	} else {
	    err = errors.New("requires --secret-keyring-b64, --secret-keyring, or SECRET_KEYRING")
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

	if err := SetValue(path, value, config); err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to load config: %v", err), 1)
	}

	return cliError(SaveConf(file, config), 1)
}
