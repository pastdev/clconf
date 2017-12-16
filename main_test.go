package main

import (
	"os"
	"path/filepath"
	"runtime"
)

func NewTestConfigFile() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "clconf", "testconfig.yml")
}

func NewTestKeysFile() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "clconf", "test.secring.gpg")
}

func WithEnv(do func()) {
	env := map[string]string{
		"YAML_FILES":     NewTestConfigFile(),
		"SECRET_KEYRING": NewTestKeysFile(),
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

func Example_noArg() {
	os.Args = []string{"clconf"}
	main()
	// Output:
	// {}
}

func Example_testConfig() {
	os.Args = []string{"clconf", "--yaml", NewTestConfigFile()}
	main()
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
	os.Args = []string{"clconf", "--yaml", NewTestConfigFile(), "getv"}
	main()
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

func Example_testConfigGetvAppAliases() {
	os.Args = []string{"clconf", "--yaml", NewTestConfigFile(), "getv", "/app/aliases"}
	main()
	// Output:
	// - foo
	// - bar
}

func Example_testConfigGetvAppDbPort() {
	os.Args = []string{"clconf", "--yaml", NewTestConfigFile(), "getv", "/app/db/port"}
	main()
	// Output:
	// 3306
}

func Example_testConfigGetvAppDbHostname() {
	os.Args = []string{
		"clconf",
		"--yaml", NewTestConfigFile(),
		"getv",
		"/app/db/hostname",
	}
	main()
	// Output:
	// db.pastdev.com
}

func Example_testConfigGetvInvalidWithDefault() {
	os.Args = []string{
		"clconf",
		"--yaml", NewTestConfigFile(),
		"getv",
		"/INVALID_PATH",
		"--default", "foo",
	}
	main()
	// Output:
	// foo
}

func Example_testConfigGetvAppDbHostnameWithDefault() {
	os.Args = []string{
		"clconf",
		"--yaml", NewTestConfigFile(),
		"getv",
		"/app/db/hostname",
		"--default", "INVALID_HOSTNAME",
	}
	main()
	// Output:
	// db.pastdev.com
}

func Example_testConfigCgetvAppDbUsername() {
	os.Args = []string{
		"clconf",
		"--yaml", NewTestConfigFile(),
		"--secret-keyring", NewTestKeysFile(),
		"cgetv",
		"/app/db/username",
	}
	main()
	// Output:
	// SECRET_USER
}

func Example_withEnvNoArg() {
	os.Args = []string{"clconf"}
	WithEnv(main)
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
	os.Args = []string{"clconf", "cgetv", "/app/db/password"}
	WithEnv(main)
	// Output:
	// SECRET_PASS
}
