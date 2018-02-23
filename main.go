package main

import (
	"os"

	"github.com/pastdev/clconf/clconf"
)

func main() {
	clconf.NewApp().Run(os.Args)
}
