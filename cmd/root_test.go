package cmd

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func Example_noArg() {
	newCmd().Execute()
	// Output:
	// {}
}

func Example_testConfig() {
	yaml := `
app:
  db:
    username: someuser
    password: somepass
    schema: appdb
    hostname: db.example.com
`
	newCmdWithYaml(yaml).Execute()
	// Output:
	// app:
	//   db:
	//     hostname: db.example.com
	//     password: somepass
	//     schema: appdb
	//     username: someuser
}

func Example_testConfigGetv() {
	yaml := `
app:
  db:
    username: someuser
    password: somepass
    schema: appdb
    hostname: db.example.com
`
	newCmdWithYaml(yaml, "getv").Execute()
	// Output:
	// app:
	//   db:
	//     hostname: db.example.com
	//     password: somepass
	//     schema: appdb
	//     username: someuser
}

func Example_testConfigGetvDecrypt() {
	newCmd(
		"--yaml", filepath.Join("..", "testdata", "testconfig.yml"),
		"--secret-keyring", filepath.Join("..", "testdata", "test.secring.gpg"),
		"getv",
		"--decrypt", "/app/db/username",
		"--decrypt", "/app/db/password",
	).Execute()
	// Output:
	// app:
	//   aliases:
	//   - foo
	//   - bar
	//   db:
	//     hostname: db.pastdev.com
	//     password: SECRET_PASS
	//     password-plaintext: SECRET_PASS
	//     port: 3306
	//     schema: clconfdb
	//     username: SECRET_USER
	//     username-plaintext: SECRET_USER
}

func Example_testConfigGetvDecryptWithPath() {
	newCmd(
		"--yaml", filepath.Join("..", "testdata", "testconfig.yml"),
		"--secret-keyring", filepath.Join("..", "testdata", "test.secring.gpg"),
		"getv",
		"/app/db",
		"--decrypt", "/username",
		"--decrypt", "/password",
	).Execute()
	// Output:
	// hostname: db.pastdev.com
	// password: SECRET_PASS
	// password-plaintext: SECRET_PASS
	// port: 3306
	// schema: clconfdb
	// username: SECRET_USER
	// username-plaintext: SECRET_USER
}

func Example_testConfigGetvDecryptWithPathAndTemplate() {
	newCmd(
		"--yaml", filepath.Join("..", "testdata", "testconfig.yml"),
		"--secret-keyring", filepath.Join("..", "testdata", "test.secring.gpg"),
		"getv",
		"/app/db",
		"--template-string", "{{ cgetv \"/username\" }}:{{ cgetv \"/password\" }}",
	).Execute()
	// Output:
	// SECRET_USER:SECRET_PASS
}

func Example_testConfigGetvDecryptWithPrefixAndPathAndTemplate() {
	newCmd(
		"--yaml", filepath.Join("..", "testdata", "testconfig.yml"),
		"--secret-keyring", filepath.Join("..", "testdata", "test.secring.gpg"),
		"--prefix", "/app/db",
		"getv",
		"/",
		"--template-string", "{{ cgetv \"/username\" }}:{{ cgetv \"/password\" }}",
	).Execute()
	// Output:
	// SECRET_USER:SECRET_PASS
}

func Example_testConfigGetvAppAliases() {
	yaml := `
app:
  db:
    username: someuser
    password: somepass
    schema: appdb
    hostname: db.example.com
    port: 3306
  aliases:
  - foo
  - bar
`
	newCmdWithYaml(yaml, "getv", "/app/aliases").Execute()
	// Output:
	// - foo
	// - bar
}

func Example_testConfigGetvAppDbPort() {
	yaml := `
app:
  db:
    username: someuser
    password: somepass
    schema: appdb
    hostname: db.example.com
    port: 3306
  aliases:
  - foo
  - bar
`
	newCmdWithYaml(yaml, "getv", "/app/db/port").Execute()
	// Output:
	// 3306
}

func Example_testConfigGetvAppDbHostname() {
	yaml := `
app:
  db:
    username: someuser
    password: somepass
    schema: appdb
    hostname: db.example.com
    port: 3306
  aliases:
  - foo
  - bar
`
	newCmdWithYaml(yaml, "getv", "/app/db/hostname").Execute()
	// Output:
	// db.example.com
}

func Example_testConfigGetvInvalidWithDefault() {
	yaml := `
app:
  db:
    username: someuser
    password: somepass
    schema: appdb
    hostname: db.example.com
    port: 3306
  aliases:
  - foo
  - bar
`
	newCmdWithYaml(yaml, "getv", "/INVALID_PATH", "--default", "foo").Execute()
	// Output:
	// foo
}

func Example_testConfigGetvAppDbHostnameWithDefault() {
	yaml := `
app:
  db:
    username: someuser
    password: somepass
    schema: appdb
    hostname: db.example.com
    port: 3306
  aliases:
  - foo
  - bar
`
	newCmdWithYaml(yaml, "getv", "/app/db/hostname", "--default", "INVALID_HOSTNAME").Execute()
	// Output:
	// db.example.com
}

func Example_testConfigCgetvAppDbUsername() {
	newCmd(
		"--yaml", filepath.Join("..", "testdata", "testconfig.yml"),
		"--secret-keyring", filepath.Join("..", "testdata", "test.secring.gpg"),
		"cgetv",
		"/app/db/username",
	).Execute()
	// Output:
	// SECRET_USER
}

func Example_withEnvNoArg() {
	WithEnv(func() { newCmd().Execute() })
	// Output:
	// app:
	//   aliases:
	//   - foo
	//   - bar
	//   db:
	//     hostname: db.pastdev.com
	//     password: wcBMA5B5A4w5Zw+rAQgALW6c2D2wwgonToJuQUmDGlnw3LG8L4dOq4qgf27L+s133trGcmBpGdsS3XysbkQ6TaYJ2y7wLpHs/dHSwrD2Z+M6WvLX5mzBhAAY5rIN+KLal7vepU+OumPGbq14kZSAYAhfkVAPxg21P04P1N/S853VPrjpeVlGWBLJMdXsGmdGLgelMAT5koSprnovsBEhm0te33KbEXSkvFVZCMF0rBwK4GV2YfPOhTwFLZCQ451Gl3fLUrdxGS6Bn9pZHl83m3lD8bFdX5kV4ezF48WREE9al3Ik/EEjcKEki2sF65mKK8a5mtEdlw8i2TzRXReUMX+QNFxNbmTyKPGpoQJ4DdLgAeS60Ee2yg9bYuB8LymvpIXe4fcj4E/gxuF9MOBb4j1cxWXg0+OcNwC7jnKTc+A04aAE4OzjvXAkVzP71PTgDuJ5DgRi4JHg3eCK4iRchCPgp+NuvJFazIksrODo5GwKh2URof5RNlbGwzLSmPvio8O96uEXYwA=
	//     password-plaintext: SECRET_PASS
	//     port: 3306
	//     schema: clconfdb
	//     username: wcBMA5B5A4w5Zw+rAQgAUfuQEe3XCfWey2j51dIl6BiDyMVcGu2nOUV+CS4GLF/AW2KfThIWICxYDEpbJhxFnGqHDkdFI8q5YowS8XDKuezJXwwkvKJkDswMiIJsHVRIoIW2kvXZHS0fJIqPN0mpUl2uPmDd+lELduV21ix4j+yO1frEgbAmKtAHvfvs5QqPOquOZVFWRnHP0SQ1Ev+argq+c1OrbSPXlGplFgfpyJWoq1vt4K2OL//us6fZtAPgNHGTIK+0hFZSTfJ7vBqEygolAO581G9fsUHWJJ+0KBj4xHy7J91mCTCCCl8gbUe6ANtSMHGcl8aNuYL6IRvOEbtZVM8MUE6MWY+k/pPABNLgAeRftcnVfmbiydJ9DXfcFePC4f364H/gcuG3AOA34mINQVng2uOpfWLop/Vv6+CE4fZy4N7jJSWyE0LgXMzgqeLRG2vc4Lvg/uAN4kxVe67gq+PSZuU8WdmEouC15LbaCnISJ/Du6cc34mhqi7DiMWHP6+EPfgA=
	//     username-plaintext: SECRET_USER
}

func Example_withEnvCgetvAppDbPassword() {
	WithEnv(func() { newCmd("cgetv", "/app/db/password").Execute() })
	// Output:
	// SECRET_PASS
}

func Example_var() {
	newCmd("var", "/foo", "bar").Execute()
	// Output:
	// /foo="bar"
}

func Example_varForceArray() {
	newCmd("var", "/foo", "bar", "--force-array").Execute()
	// Output:
	// /foo=["bar"]
}

func Example_varValueOnly() {
	newCmd("var", "/foo", "bar", "--value-only").Execute()
	// Output:
	// "bar"
}

func Example_varForceArrayValueOnly() {
	newCmd("var", "/foo", "bar", "--force-array", "--value-only").Execute()
	// Output:
	// ["bar"]
}

func Example_varArray() {
	newCmd("var", "/foo", "bar", "baz").Execute()
	// Output:
	// /foo=["bar","baz"]
}

func Example_getvVar() {
	newCmd("getv", "--var", `/foo="bar"`).Execute()
	// Output:
	// foo: bar
}

func Example_getvMultipleVar() {
	newCmd("getv", "--var", `/foo/baz="bar"`, "--var", `/foo/hip="hop"`).Execute()
	// Output:
	// foo:
	//   baz: bar
	//   hip: hop
}

func Example_getvObject() {
	newCmd("getv", "--var", `/={"foo":{"baz":"bar","hip":"hop"}}`).Execute()
	// Output:
	// foo:
	//   baz: bar
	//   hip: hop
}

func Example_getvStringAsJson() {
	newCmd("getv", "--var", `/foo="bar"`, "/foo", "--as-json").Execute()
	// Output:
	// "bar"
}

func Example_getvScalarAsBashArray() {
	newCmd("getv", "--var", `/a="bar"`, "/a", "--as-bash-array").Execute()
	// Output:
	// ([0]="bar")
}

func Example_getvArrayAsBashArray() {
	newCmd("getv", "--var", `/a=["foo","bar"]`, "/a", "--as-bash-array").Execute()
	// Output:
	// ([0]="foo" [1]="bar")
}

func Example_getvMapAsBashArray() {
	newCmd("getv", "--var", `/a={"foo":"bar","hip":"hop"}`, "/a", "--as-bash-array").Execute()
	// Output:
	// ([0]="{\"key\":\"foo\",\"value\":\"bar\"}" [1]="{\"key\":\"hip\",\"value\":\"hop\"}")
}

func Example_getvArrayOfObjectsAsBashArray() {
	newCmd("getv", "--var", `/a=[{"foo":"bar"},{"hip":"hop"}]`, "/a", "--as-bash-array").Execute()
	// Output:
	// ([0]="{\"foo\":\"bar\"}" [1]="{\"hip\":\"hop\"}")
}

func Example_mergeOverridesBooleanToFalse() {
	footrue, err := ioutil.TempFile("", "")
	if err != nil {
		return
	}
	defer os.Remove(footrue.Name())

	if _, err := footrue.Write([]byte("---\nfoo: true")); err != nil {
		return
	}

	foofalse, err := ioutil.TempFile("", "")
	if err != nil {
		return
	}
	defer os.Remove(foofalse.Name())

	if _, err := foofalse.Write([]byte("---\nfoo: false")); err != nil {
		return
	}

	newCmd("getv", "--yaml", footrue.Name(), "--yaml", foofalse.Name()).Execute()
	// Output:
	// foo: false
}

func Example_preserveListOrderInRange() {
	yaml := `
a_list:
- zebra
- elephant
- cat
- unicorn
`
	newCmdWithYaml(yaml, "getv", "/", "--template-string",
		`{{ range (getksvs "/a_list/*" "int") }}{{.}}{{"\n"}}{{ end }}`).Execute()
	// Output:
	// zebra
	// elephant
	// cat
	// unicorn
}

func Example_sortListInRange() {
	yaml := `
a_list:
- zebra
- elephant
- cat
- unicorn
`
	newCmdWithYaml(yaml, "getv", "/", "--template-string",
		`{{ range getsvs "/a_list/*" }}{{.}}{{"\n"}}{{ end }}`).Execute()
	// Output:
	// cat
	// elephant
	// unicorn
	// zebra
}

func newCmd(args ...string) *cobra.Command {
	cmd := rootCmd()
	cmd.SetArgs(args)
	return cmd
}

func newCmdWithYaml(yaml string, args ...string) *cobra.Command {
	b64yml := base64.StdEncoding.EncodeToString([]byte(yaml))
	cmd := rootCmd()
	args = append([]string{"--yaml-base64", b64yml}, args...)
	cmd.SetArgs(args)
	return cmd
}

func WithEnv(do func()) {
	env := map[string]string{
		"YAML_FILES":     filepath.Join("..", "testdata", "testconfig.yml"),
		"SECRET_KEYRING": filepath.Join("..", "testdata", "test.secring.gpg"),
	}
	defer func() {
		for key := range env {
			os.Unsetenv(key)
		}
	}()
	for key, value := range env {
		os.Setenv(key, value)
	}
	do()
}
