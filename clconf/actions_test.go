package clconf

import (
	"encoding/base64"
	"flag"
	"testing"

	"github.com/urfave/cli"
)

const testConfigYaml = "" +
	"app:\n" +
	"  db:\n" +
	"    hostname: db.pastdev.com\n" +
	"    password: SECRET_PASS\n" +
	"    schema: clconfdb\n" +
	"    username: SECRET_USER\n"

func testGlobalContext() *cli.Context {
	set := flag.NewFlagSet("test", 0)
	set.String("override", base64.StdEncoding.EncodeToString([]byte(testConfigYaml)), "override data")
	set.String("secret-keys-file", testSecretKeysFile(), "secret kesy file")
	return cli.NewContext(nil, set, nil)
}

func TestGetv(t *testing.T) {
	actual := getv(cli.NewContext(nil, nil, testGlobalContext()))
}
