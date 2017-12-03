package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"
	"gitlab.com/pastdev/s2i/clconf/clconf"
)

// Thoughts...
// * there should be actions `getv`, `getc`, `setv`, `setc`
// ** all take a coordinate, setters also take a _fileish_ thing
// *** setter _fileish_ thing could be cached as env var...
// ** `getv` should allow for a --decrypt option which takes a list of coords to decrypt
// * --override opition (multi valued) takes base64 encoded yaml
// ** stdin should be read in as a file and used as override (not sure if before or after --overrides)
func main() {
	app := cli.NewApp()
	app.Name = "clconf"
	app.Action = func(c *cli.Context) error {
		yaml, err := clconf.MarshalYaml(clconf.LoadConf())
		if err != nil {
			log.Fatalf("Unable to load conf: %v", err)
		}
		fmt.Println(yaml)
		return nil
	}

	app.Run(os.Args)
}
