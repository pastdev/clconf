package main

import (
	"os"
	"regexp"

	"gitlab.com/pastdev/s2i/clconf/clconf"
)

var splitter = regexp.MustCompile(`\s+`)

// YAML_CONFIGS=[{type:secret, path:/foo/bar}, {path:/not/a/secret}, {type: secret, path:/hip/hop}]
// CONFIG_PREFIX=
// duplicates for all env vars as command line args.
func main() {
	yamls := append([]string{},
		clconf.ReadFiles(
			splitter.Split(os.Getenv("YAML_FILES"), -1)...)...)
	yamls = append(yamls, 
		clconf.DecodeBase64Strings(
			clconf.ReadEnvVars(
                splitter.Split(os.Getenv("YAML_VARS"), -1)...)...)...)
}
