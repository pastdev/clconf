package cmd

import (
	"os"
	"path/filepath"
)

func Example_noArg() {
	reinit()
	os.Args = []string{"clconf"}
	Execute()
	// Output:
	// {}
}

func Example_testConfig() {
	reinit()
	os.Args = []string{
		"clconf",
		"--yaml", filepath.Join("..", "testdata", "testconfig.yml"),
	}
	Execute()
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

func Example_testConfigGetv() {
	reinit()
	os.Args = []string{
		"clconf",
		"--yaml", filepath.Join("..", "testdata", "testconfig.yml"),
		"getv",
	}
	Execute()
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

func Example_testConfigGetvDecrypt() {
	reinit()
	os.Args = []string{
		"clconf",
		"--yaml", filepath.Join("..", "testdata", "testconfig.yml"),
		"--secret-keyring", filepath.Join("..", "testdata", "test.secring.gpg"),
		"getv",
		"--decrypt", "/app/db/username",
		"--decrypt", "/app/db/password",
	}
	Execute()
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
	reinit()
	os.Args = []string{
		"clconf",
		"--yaml", filepath.Join("..", "testdata", "testconfig.yml"),
		"--secret-keyring", filepath.Join("..", "testdata", "test.secring.gpg"),
		"getv",
		"/app/db",
		"--decrypt", "/username",
		"--decrypt", "/password",
	}
	Execute()
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
	reinit()
	os.Args = []string{
		"clconf",
		"--yaml", filepath.Join("..", "testdata", "testconfig.yml"),
		"--secret-keyring", filepath.Join("..", "testdata", "test.secring.gpg"),
		"getv",
		"/app/db",
		"--template-string", "{{ cgetv \"/username\" }}:{{ cgetv \"/password\" }}",
	}
	Execute()
	// Output:
	// SECRET_USER:SECRET_PASS
}

func Example_testConfigGetvDecryptWithPrefixAndPathAndTemplate() {
	reinit()
	os.Args = []string{
		"clconf",
		"--yaml", filepath.Join("..", "testdata", "testconfig.yml"),
		"--secret-keyring", filepath.Join("..", "testdata", "test.secring.gpg"),
		"--prefix", "/app/db",
		"getv",
		"/",
		"--template-string", "{{ cgetv \"/username\" }}:{{ cgetv \"/password\" }}",
	}
	Execute()
	// Output:
	// SECRET_USER:SECRET_PASS
}

func Example_testConfigGetvAppAliases() {
	reinit()
	os.Args = []string{
		"clconf",
		"--yaml", filepath.Join("..", "testdata", "testconfig.yml"),
		"--secret-keyring", filepath.Join("..", "testdata", "test.secring.gpg"),
		"getv",
		"/app/aliases",
	}
	Execute()
	// Output:
	// - foo
	// - bar
}

func Example_testConfigGetvAppDbPort() {
	reinit()
	os.Args = []string{
		"clconf",
		"--yaml", filepath.Join("..", "testdata", "testconfig.yml"),
		"getv",
		"/app/db/port",
	}
	Execute()
	// Output:
	// 3306
}

func Example_testConfigGetvAppDbHostname() {
	reinit()
	os.Args = []string{
		"clconf",
		"--yaml", filepath.Join("..", "testdata", "testconfig.yml"),
		"getv",
		"/app/db/hostname",
	}
	Execute()
	// Output:
	// db.pastdev.com
}

func Example_testConfigGetvInvalidWithDefault() {
	reinit()
	os.Args = []string{
		"clconf",
		"--yaml", filepath.Join("..", "testdata", "testconfig.yml"),
		"getv",
		"/INVALID_PATH",
		"--default", "foo",
	}
	Execute()
	// Output:
	// foo
}

func Example_testConfigGetvAppDbHostnameWithDefault() {
	reinit()
	os.Args = []string{
		"clconf",
		"--yaml", filepath.Join("..", "testdata", "testconfig.yml"),
		"--secret-keyring", filepath.Join("..", "testdata", "test.secring.gpg"),
		"getv",
		"/app/db/hostname",
		"--default", "INVALID_HOSTNAME",
	}
	Execute()
	// Output:
	// db.pastdev.com
}

func Example_testConfigCgetvAppDbUsername() {
	reinit()
	os.Args = []string{
		"clconf",
		"--yaml", filepath.Join("..", "testdata", "testconfig.yml"),
		"--secret-keyring", filepath.Join("..", "testdata", "test.secring.gpg"),
		"cgetv",
		"/app/db/username",
	}
	Execute()
	// Output:
	// SECRET_USER
}

func Example_withEnvNoArg() {
	reinit()
	os.Args = []string{"clconf"}
	WithEnv(Execute)
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
	reinit()
	os.Args = []string{"clconf", "cgetv", "/app/db/password"}
	WithEnv(Execute)
	// Output:
	// SECRET_PASS
}

func reinit() {
	// package variables are set at beginning of run, and not re-set
	// between tests.  we must do so manually
	rootCmdContext.prefix = *newOptionalString("", false)
	rootCmdContext.secretKeyring = *newOptionalString("", false)
	rootCmdContext.secretKeyringBase64 = *newOptionalString("", false)
	rootCmdContext.yaml = []string{}
	rootCmdContext.yamlBase64 = []string{}
	getvCmdContext.decrypt = []string{}
	getvCmdContext.defaultValue = *newOptionalString("", false)
	getvCmdContext.template = *newOptionalString("", false)
	getvCmdContext.templateBase64 = *newOptionalString("", false)
	getvCmdContext.templateString = *newOptionalString("", false)
	setvCmdContext.encrypt = false
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
