package main

import (
	"os"

	"github.com/urfave/cli"
	"gitlab.com/pastdev/s2i/clconf/clconf"
)

const name = "clconf"

// Thoughts...
// * default action is getv
// ** getv default coord is /
// * there should be actions `getv`, `cgetv`, `setv`, `csetv`
// ** all take a coordinate, setters also take a _fileish_ thing
// *** setter _fileish_ thing could be cached as env var...
// ** `getv` should allow for a --decrypt option which takes a list of coords to decrypt
// * --override opition (multi valued) takes base64 encoded yaml
// ** stdin should be read in as a file and used as override (not sure if before or after --overrides)
func main() {
	app := cli.NewApp()
	app.Name = name
	app.Version = "0.0.1"
	app.UsageText = "clconf [global options] command [command options] [args...]"

	app.Flags = []cli.Flag {
		cli.StringSliceFlag{Name: "yaml-file"},
		cli.StringSliceFlag{Name: "yaml-vars"},
		cli.StringSliceFlag{Name: "override"},
		cli.StringFlag{Name: "secret-keys"},
		cli.StringFlag{Name: "secret-keys-file"},
	}

    app.Commands = []cli.Command{
		{
			Name: "cgetv",
			Usage: "Get a secret value",
			ArgsUsage: "PATH",
			Action: clconf.Cgetv,
		},
		{
			Name: "getv",
			Usage: "Get a value",
			ArgsUsage: "PATH",
			Action: clconf.Getv,
		},
	}

	app.Action = clconf.Getv

	app.Run(os.Args)
}
