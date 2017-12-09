package clconf

import (
	"fmt"

	"github.com/urfave/cli"
)

const (
	// Name is the name of this cli application
	Name = "clconf"
)

// https://stackoverflow.com/a/26804949/516433
var print = fmt.Print

func cgetv(c *cli.Context) error {
	return dump(marshal(cgetvHandler(c)))
}

func cgetvHandler(c *cli.Context) (*cli.Context, interface{}, error) {
	path := c.Args().First()
	value, ok := GetValue(path, load(c))
	if !ok {
		return c, nil, cli.NewExitError(fmt.Sprintf("[%v] does not exist", path), 1)
	}
	secretAgent, err := newSecretAgentFromCli(c)
	if err != nil {
		return c, nil, err
	}
	if stringValue, ok := value.(string); ok {
	    decrypted, err := secretAgent.Decrypt(stringValue)
	    return c, decrypted, err
	}
	return c, nil, cli.NewExitError(fmt.Sprintf("[%v] not a string value", path), 1)
}

func marshal(c *cli.Context, value interface{}, err error) (*cli.Context, string, error) {
	if err != nil {
		return c, "", err
	}
	if stringValue, ok := value.(string); ok {
		return c, stringValue, err
	} else if mapValue, ok := value.(map[interface{}]interface{}); ok {
		marshaled, err := MarshalYaml(mapValue)
		return c, marshaled, err
	} else if arrayValue, ok := value.([]interface{}); ok {
		marshaled, err := MarshalYaml(arrayValue)
		return c, marshaled, err
	}
	return c, fmt.Sprintf("%v", value), err
}

func dump(c *cli.Context, value interface{}, err error) cli.ExitCoder {
	if err != nil {
		if casted, ok := err.(cli.ExitCoder); ok {
			return casted
		} 
		return cli.NewExitError(err, 1)
	}
	print(value)
	return nil
}

func getv(c *cli.Context) error {
	return dump(marshal(getvHandler(c)))
}

func getvFlags() []cli.Flag {
	return []cli.Flag{}
}

func getvHandler(c *cli.Context) (*cli.Context, interface{}, error) {
	path := c.Args().First()
	if value, ok := GetValue(path, load(c)); ok {
		return c, value, nil
	}
	return c, nil, cli.NewExitError(
		fmt.Sprintf("[%v] does not exist", path), 1)
}

func globalFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringSliceFlag{Name: "yaml-file"},
		cli.StringSliceFlag{Name: "override"},
		cli.StringFlag{Name: "secret-keys"},
		cli.StringFlag{Name: "secret-keys-file"},
	}
}

func load(c *cli.Context) map[interface{}]interface{} {
	return LoadConfFromEnvironment(
		c.GlobalStringSlice("yaml-file"),
		c.GlobalStringSlice("override"))
}

// NewApp returns a new cli application instance ready to be run.
// 
// Thoughts...
// * default action is getv
// ** getv default coord is /
// * there should be actions `getv`, `cgetv`, `setv`, `csetv`
// ** all take a coordinate, setters also take a _fileish_ thing
// *** setter _fileish_ thing could be cached as env var...
// ** `getv` should allow for a --decrypt option which takes a list of coords to decrypt
// * --override opition (multi valued) takes base64 encoded yaml
// ** stdin should be read in as a file and used as override (not sure if before or after --overrides)
func NewApp() *cli.App {
	app := cli.NewApp()
	app.Name = Name
	app.Version = "0.0.1"
	app.UsageText = "clconf [global options] command [command options] [args...]"

	app.Flags = globalFlags()

    app.Commands = []cli.Command{
		{
			Name: "cgetv",
			Usage: "Get a secret value",
			ArgsUsage: "PATH",
			Action: cgetv,
		},
		{
			Name: "getv",
			Usage: "Get a value",
			ArgsUsage: "PATH",
			Action: getv,
		},
	}

	app.Action = getv

	return app
}

func newSecretAgentFromCli(c *cli.Context) (*SecretAgent, error) {
	if secretKeysBase64 := c.GlobalString("secret-keys"); secretKeysBase64 != "" {
		return NewSecretAgentFromBase64(secretKeysBase64)
	}
	if secretKeysFile := c.GlobalString("secret-keys-file"); secretKeysFile != "" {
		return NewSecretAgentFromFile(secretKeysFile)
	}
	return nil, cli.NewExitError("--secret-keys or --secret-keys-file required", 1)
}
