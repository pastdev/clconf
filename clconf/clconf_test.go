package clconf_test

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/imdario/mergo"
	"github.com/mitchellh/mapstructure"
	"github.com/pastdev/clconf/v2/clconf"
)

const configMap = "" +
	"---\n" +
	"key: foobar\n" +
	"database:\n" +
	"  host: 127.0.0.1\n" +
	"  password: REDACTED\n" +
	"  port: \"3306\"\n" +
	"  username: REDACTED\n" +
	"upstream:\n" +
	"  app1: 10.0.1.10:8080\n" +
	"  app2: 10.0.1.11:8080\n" +
	"prefix:\n" +
	"  database:\n" +
	"    host: 127.0.0.1\n" +
	"    password: REDACTED\n" +
	"    port: \"3306\"\n" +
	"    username: REDACTED\n" +
	"  upstream:\n" +
	"    app1: 10.0.1.10:8080\n" +
	"    app2: 10.0.1.11:8080\n"
const secrets = "" +
	"---\n" +
	"database:\n" +
	"  password: p@sSw0rd\n" +
	"  username: confd\n" +
	"prefix:\n" +
	"  database:\n" +
	"    password: p@sSw0rd\n" +
	"    username: confd\n"
const yaml1and2 = "" +
	"a: Yup\n" +
	"b:\n" +
	"  c: 2\n" +
	"  e: 2\n" +
	"  f:\n" +
	"    g: foobar\n"
const yamlWithList = "" +
	"a:\n" +
	"- one\n" +
	"- two\n" +
	"- three\n"

var configMapAndSecretsExpected = map[interface{}]interface{}{
	"key": "foobar",
	"database": map[interface{}]interface{}{
		"host":     "127.0.0.1",
		"password": "p@sSw0rd",
		"port":     "3306",
		"username": "confd",
	},
	"upstream": map[interface{}]interface{}{
		"app1": "10.0.1.10:8080",
		"app2": "10.0.1.11:8080",
	},
	"prefix": map[interface{}]interface{}{
		"database": map[interface{}]interface{}{
			"host":     "127.0.0.1",
			"password": "p@sSw0rd",
			"port":     "3306",
			"username": "confd",
		},
		"upstream": map[interface{}]interface{}{
			"app1": "10.0.1.10:8080",
			"app2": "10.0.1.11:8080",
		},
	},
}

func assertMergeValue(
	t *testing.T,
	expected interface{},
	conf interface{},
	path string,
	value interface{},
	overwrite bool,
) {
	expectedYaml, _ := clconf.MarshalYaml(expected)
	confYaml, _ := clconf.MarshalYaml(conf)
	valueYaml, _ := clconf.MarshalYaml(value)
	err := clconf.MergeValue(conf, path, value, overwrite)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if !reflect.DeepEqual(expected, conf) {
		actualYaml, _ := clconf.MarshalYaml(conf)
		t.Errorf(
			"\nexpected:\n---\n%v\nactual:\n---\n%v\nconf:\n---\n%v\nvalue at %s\n---\n%v",
			string(expectedYaml),
			string(actualYaml),
			string(confYaml),
			path,
			string(valueYaml))
	}
}

func TestBase64Strings(t *testing.T) {
	encoded := []string{}
	actual, err := clconf.DecodeBase64Strings(encoded...)
	if err != nil || len(actual) != 0 {
		t.Errorf("Base64Strings empty failed: [%v]", actual)
	}

	expected := []string{"one", "two"}
	encoded = []string{
		base64.StdEncoding.EncodeToString([]byte(expected[0])),
		base64.StdEncoding.EncodeToString([]byte(expected[1]))}
	actual, err = clconf.DecodeBase64Strings(encoded...)
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("Base64Strings one two failed: [%v] == [%v]", expected, actual)
	}

	if _, err := clconf.DecodeBase64Strings("&*INVALID*&"); err == nil {
		t.Error("Base64Strings invalid should have failed")
	}
}

func TestFill(t *testing.T) {
	expectedTimestampString := "2019-10-16T12:28:49Z"
	conf, err := clconf.UnmarshalYaml("---\n" +
		"data:\n" +
		fmt.Sprintf("  timestamp: %s\n", expectedTimestampString) +
		"  bool_as_string: 'true'\n" +
		"  bool_as_bool: false\n" +
		"  string_value: 'foo'\n" +
		"  subtype:\n" +
		"    value: 'bar'\n")
	if err != nil {
		t.Fatalf("unmarshal yaml failed: %v", err)
	}

	type Subtype struct {
		Value string
	}

	type Data struct {
		Timestamp    time.Time
		BoolAsString bool   `yaml:"bool_as_string"`
		BoolAsBool   bool   `yaml:"bool_as_bool"`
		StringValue  string `yaml:"string_value"`
		Subtype      Subtype
	}

	type Root struct {
		Data Data
	}

	actual := Root{}
	err = clconf.Fill(
		"",
		conf,
		&mapstructure.DecoderConfig{
			DecodeHook:       mapstructure.StringToTimeHookFunc(time.RFC3339),
			Result:           &actual,
			TagName:          "yaml",
			WeaklyTypedInput: true,
		})
	if err != nil {
		t.Fatalf("unable to fill: %v", err)
	}

	expectedTimestamp, err := time.Parse(
		time.RFC3339,
		expectedTimestampString)
	if err != nil {
		t.Fatalf("unable to parse [%s]: %v", expectedTimestampString, err)
	}
	expected := Root{
		Data{
			Timestamp:    expectedTimestamp,
			BoolAsString: true,
			BoolAsBool:   false,
			StringValue:  "foo",
			Subtype: Subtype{
				Value: "bar",
			},
		},
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("%v != %v", expected, actual)
	}
}

func TestFillValue(t *testing.T) {
	conf, _ := clconf.UnmarshalYaml(yaml1and2)

	type Eff struct {
		G string
	}
	type Bee struct {
		C int
		E int
		F Eff
	}
	type Root struct {
		A string
		B Bee
	}
	var root Root
	ok := clconf.FillValue("", conf, &root)
	expectedRoot := Root{A: "Yup", B: Bee{C: 2, E: 2, F: Eff{G: "foobar"}}}
	if !ok || !reflect.DeepEqual(expectedRoot, root) {
		t.Errorf("FillValue empty path failed: [%v] [%v] == [%v]", ok, expectedRoot, root)
	}

	var bee Bee
	ok = clconf.FillValue("b", conf, &bee)
	expectedBee := Bee{C: 2, E: 2, F: Eff{G: "foobar"}}
	if !ok || !reflect.DeepEqual(expectedBee, bee) {
		t.Errorf("FillValue first level failed: [%v] [%v] == [%v]", ok, expectedBee, bee)
	}

	type BeeLight struct {
		C int
		E int
	}
	var beeLight BeeLight
	ok = clconf.FillValue("b", conf, &beeLight)
	expectedBeeLight := BeeLight{C: 2, E: 2}
	if !ok || !reflect.DeepEqual(expectedBeeLight, beeLight) {
		t.Errorf("FillValue first level string failed: [%v] [%v] == [%v]", ok, expectedBeeLight, beeLight)
	}

	type Zee struct {
		ShouldntWork string
	}
	var zee Zee
	ok = clconf.FillValue("/z", conf, &zee)
	if ok {
		t.Error("FillValue invalid path should have failed")
	}

	ok = clconf.FillValue("a/", conf, &zee)
	if ok {
		t.Error("FillValue a should not have been z")
	}
}

func TestGetValue(t *testing.T) {
	conf, _ := clconf.UnmarshalYaml(yaml1and2)

	value, err := clconf.GetValue(conf, "")
	if err != nil || !reflect.DeepEqual(conf, value) {
		t.Errorf("GetValue empty path failed: [%v] [%v] == [%v]", err, conf, value)
	}

	value, err = clconf.GetValue(conf, "/")
	if err != nil || !reflect.DeepEqual(conf, value) {
		t.Errorf("GetValue empty path failed: [%v] [%v] == [%v]", err, conf, value)
	}

	value, err = clconf.GetValue(conf, "/a")
	if err != nil || value != "Yup" {
		t.Errorf("GetValue first level string failed: [%v] [%v]", err, value)
	}

	value, err = clconf.GetValue(conf, "/b//f//g")
	if err != nil || value != "foobar" {
		t.Errorf("GetValue third level string multi slash failed: [%v] [%v]", err, value)
	}

	value, err = clconf.GetValue(conf, "/a/f")
	if err == nil {
		t.Errorf("GetValue non map indexing should have failed: [%v] [%v]", err, value)
	}

	value, err = clconf.GetValue(conf, "/z")
	if err == nil {
		t.Errorf("GetValue missing have failed: [%v] [%v]", err, value)
	}

	conf, _ = clconf.UnmarshalYaml(yamlWithList)

	value, err = clconf.GetValue(conf, "/a")
	expected := []interface{}{"one", "two", "three"}
	if err != nil || !reflect.DeepEqual(expected, value) {
		t.Errorf("GetValue list failed: (err:[%v]) [%v] == [%v]", err, expected, value)
	}

	value, err = clconf.GetValue(conf, "/a/0")
	stringExpected := "one"
	if err != nil || !reflect.DeepEqual(stringExpected, value) {
		t.Errorf("GetValue list item failed: (err:[%v]) [%v] == [%v]", err, stringExpected, value)
	}

	_, err = clconf.GetValue(conf, "/a/b")
	if err == nil {
		t.Errorf("GetValue list item invalid index should have failed")
	}
}

func TestLoadConf(t *testing.T) {
	envVars := []string{"a"}
	tempDir, err := ioutil.TempDir("", "clconf")
	if err != nil {
		t.Errorf("Unable to create temp dir: %v", err)
	}
	defer func() {
		os.RemoveAll(tempDir)
		for _, name := range envVars {
			os.Unsetenv(name)
		}
		os.Unsetenv("YAML_FILES")
		os.Unsetenv("YAML_VARS")
		os.Unsetenv("YAML_VAR")
	}()

	envValues := []string{base64.StdEncoding.EncodeToString([]byte("a: env\nenv: env"))}
	for index, name := range envVars {
		os.Setenv(name, envValues[index])
	}

	stdinFile := path.Join(tempDir, "stdin")
	err = ioutil.WriteFile(stdinFile, []byte("a: stdin\nstdin: 1"), 0700)
	if err != nil {
		t.Errorf("failed to write stdinFile: %v", err)
	}

	fileArg := path.Join(tempDir, "fileArg")
	err = ioutil.WriteFile(fileArg, []byte("a: fileArg\nfileArg: 1"), 0700)
	if err != nil {
		t.Errorf("failed to write fileArg: %v", err)
	}
	fileEnv := path.Join(tempDir, "fileEnv")
	err = ioutil.WriteFile(fileEnv, []byte("a: fileEnv\nfileEnv: 1"), 0700)
	if err != nil {
		t.Errorf("failed to write fileEnv: %v", err)
	}

	b64Arg := base64.StdEncoding.EncodeToString([]byte("a: b64Arg\nb64Arg: 1"))
	b64Env := base64.StdEncoding.EncodeToString([]byte("a: b64Env\nb64Env: 1"))

	actual, err := clconf.LoadConf([]string{}, []string{})
	if err != nil || len(actual) > 0 {
		t.Errorf("LoadConf no config failed")
	}

	expected, _ := clconf.UnmarshalYaml("a: b64Arg\nb64Arg: 1")
	actual, err = clconf.LoadConfFromEnvironment([]string{}, []string{b64Arg})
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf b64Arg failed: [%v] != [%v] (err: %v)", expected, actual, err)
	}

	os.Setenv("YAML_VAR", b64Env)
	os.Setenv("YAML_VARS", "YAML_VAR")
	expected, _ = clconf.UnmarshalYaml("a: b64Env\nb64Arg: 1\nb64Env: 1")
	actual, err = clconf.LoadConfFromEnvironment([]string{}, []string{b64Arg})
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf b64Arg, b64Env failed: [%v] != [%v] (err: %v)", expected, actual, err)
	}
	os.Unsetenv("YAML_VARS")
	os.Unsetenv("YAML_VAR")

	os.Setenv("YAML_VAR", b64Env)
	os.Setenv("YAML_VARS", "YAML_VAR")
	expected, _ = clconf.UnmarshalYaml("a: b64Arg\nb64Arg: 1")
	actual, err = clconf.LoadConf([]string{}, []string{b64Arg})
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf b64Arg, b64Env (but disabled) failed: [%v] != [%v] (err: %v)", expected, actual, err)
	}
	os.Unsetenv("YAML_VARS")
	os.Unsetenv("YAML_VAR")

	expected, _ = clconf.UnmarshalYaml("a: fileArg\nfileArg: 1")
	actual, err = clconf.LoadConfFromEnvironment([]string{fileArg}, []string{})
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf fileArg failed: [%v] != [%v] (err: %v)", expected, actual, err)
	}

	os.Setenv("YAML_FILES", fileEnv)
	expected, _ = clconf.UnmarshalYaml("a: fileEnv\nfileArg: 1\nfileEnv: 1")
	actual, err = clconf.LoadConfFromEnvironment([]string{fileArg}, []string{})
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf fileArg, fileEnv failed: [%v] != [%v] (err: %v)", expected, actual, err)
	}
	os.Unsetenv("YAML_FILES")

	os.Setenv("YAML_VAR", b64Env)
	os.Setenv("YAML_VARS", "YAML_VAR")
	os.Setenv("YAML_FILES", fileEnv)
	expected, _ = clconf.UnmarshalYaml("a: b64Env\nfileArg: 1\nfileEnv: 1\nb64Arg: 1\nb64Env: 1")
	actual, err = clconf.LoadConfFromEnvironment([]string{fileArg}, []string{b64Arg})
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf fileArg, fileEnv, b64Arg, b64Env failed: [%v] != [%v] (err: %v)", expected, actual, err)
	}
	os.Unsetenv("YAML_FILES")
	os.Unsetenv("YAML_VARS")
	os.Unsetenv("YAML_VAR")

	stdin, err := os.Open(stdinFile)
	if err != nil {
		t.Errorf("Error opening stdin file for reading")
	}
	defer stdin.Close()
	os.Setenv("YAML_VAR", b64Env)
	os.Setenv("YAML_VARS", "YAML_VAR")
	os.Setenv("YAML_FILES", fileEnv)
	expected, _ = clconf.UnmarshalYaml("a: stdin\nfileArg: 1\nfileEnv: 1\nb64Arg: 1\nb64Env: 1\nstdin: 1")
	actual, err = clconf.ConfSources{
		Files:       []string{fileArg},
		Overrides:   []string{b64Arg},
		Environment: true,
		Stream:      stdin}.Load()
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf all failed: [%v] != [%v] (err: %v)", expected, actual, err)
	}
	os.Unsetenv("YAML_FILES")
	os.Unsetenv("YAML_VARS")
	os.Unsetenv("YAML_VAR")
}

func TestMarshalYaml(t *testing.T) {
	value := map[interface{}]interface{}{"a": "b"}
	yaml, err := clconf.MarshalYaml(value)
	if err != nil || string(yaml) != "a: b\n" {
		t.Errorf("Marshal failed for [%v]: [%v] [%v]", value, err, yaml)
	}
}

func TestMerge(t *testing.T) {
	result := make(map[interface{}]interface{})

	configMap := map[interface{}]interface{}{
		"foo": "bar",
		"database": map[interface{}]interface{}{
			"hostname": "localhost",
			"port":     3306,
			"username": "admin",
		},
	}
	secrets := map[interface{}]interface{}{
		"hip": "hop",
		"database": map[interface{}]interface{}{
			"password": "p@ssw0rD",
			"username": "notadmin",
		},
	}

	if err := mergo.Merge(&result, secrets); err != nil {
		t.Errorf("merge failed: [%v]", err)
	}
	if !reflect.DeepEqual(result, secrets) {
		t.Errorf("merge incorrect: [%v] != [%v]", result, secrets)
	}

	expected := map[interface{}]interface{}{
		"foo": "bar",
		"hip": "hop",
		"database": map[interface{}]interface{}{
			"hostname": "localhost",
			"password": "p@ssw0rD",
			"port":     3306,
			"username": "notadmin",
		},
	}

	if err := mergo.Merge(&result, configMap); err != nil {
		t.Errorf("merge failed: [%v]", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("merge incorrect: [%v] != [%v]", result, expected)
	}
}

func TestMergeValue(t *testing.T) {
	t.Run("merge root",
		func(t *testing.T) {
			conf := map[interface{}]interface{}{"foo": "baz"}
			value := map[interface{}]interface{}{"foo": "bar"}
			assertMergeValue(t,
				map[interface{}]interface{}{"foo": "baz"},
				conf,
				"/",
				value,
				false)
		})
	t.Run("merge root overwrite",
		func(t *testing.T) {
			conf := map[interface{}]interface{}{"foo": "baz"}
			value := map[interface{}]interface{}{"foo": "bar"}
			assertMergeValue(t,
				map[interface{}]interface{}{"foo": "bar"},
				conf,
				"/",
				value,
				true)
		})
	t.Run("merge root deep",
		func(t *testing.T) {
			conf := map[interface{}]interface{}{
				"foo": map[interface{}]interface{}{
					"bar": "bap",
					"hip": "hop",
				},
			}
			value := map[interface{}]interface{}{
				"foo": map[interface{}]interface{}{
					"bar": "baz",
					"zip": "zap",
				},
			}
			assertMergeValue(t,
				map[interface{}]interface{}{
					"foo": map[interface{}]interface{}{
						"bar": "bap",
						"hip": "hop",
						"zip": "zap",
					},
				},
				conf,
				"/",
				value,
				false)
		})
	t.Run("merge root deep overwrite",
		func(t *testing.T) {
			conf := map[interface{}]interface{}{
				"foo": map[interface{}]interface{}{
					"bar": "bap",
					"hip": "hop",
				},
			}
			value := map[interface{}]interface{}{
				"foo": map[interface{}]interface{}{
					"bar": "baz",
					"zip": "zap",
				},
			}
			assertMergeValue(t,
				map[interface{}]interface{}{
					"foo": map[interface{}]interface{}{
						"bar": "baz",
						"hip": "hop",
						"zip": "zap",
					},
				},
				conf,
				"/",
				value,
				true)
		})
	t.Run("merge deep",
		func(t *testing.T) {
			conf := map[interface{}]interface{}{
				"foo": map[interface{}]interface{}{
					"bar": "bap",
					"hip": "hop",
				},
			}
			value := map[interface{}]interface{}{
				"bar": "baz",
				"zip": "zap",
			}
			assertMergeValue(t,
				map[interface{}]interface{}{
					"foo": map[interface{}]interface{}{
						"bar": "bap",
						"hip": "hop",
						"zip": "zap",
					},
				},
				conf,
				"/foo",
				value,
				false)
		})
	t.Run("merge deep overwrite",
		func(t *testing.T) {
			conf := map[interface{}]interface{}{
				"foo": map[interface{}]interface{}{
					"bar": "bap",
					"hip": "hop",
				},
			}
			value := map[interface{}]interface{}{
				"bar": "baz",
				"zip": "zap",
			}
			assertMergeValue(t,
				map[interface{}]interface{}{
					"foo": map[interface{}]interface{}{
						"bar": "baz",
						"hip": "hop",
						"zip": "zap",
					},
				},
				conf,
				"/foo",
				value,
				true)
		})
	t.Run("merge bool false override",
		func(t *testing.T) {
			conf := map[interface{}]interface{}{
				"sub": map[interface{}]interface{}{"foo": true, "bar": 456}}
			value := map[interface{}]interface{}{
				"sub": map[interface{}]interface{}{"foo": false, "bar": 123}}
			assertMergeValue(t,
				map[interface{}]interface{}{
					"sub": map[interface{}]interface{}{"foo": false, "bar": 123}},
				conf,
				"/",
				value,
				true)
		})
	t.Run("merge bool true override",
		func(t *testing.T) {
			conf := map[interface{}]interface{}{"foo": false, "bar": 456}
			value := map[interface{}]interface{}{"foo": true, "bar": 123}
			assertMergeValue(t,
				map[interface{}]interface{}{"foo": true, "bar": 123},
				conf,
				"/",
				value,
				true)
		})
}

func TestReadEnvVars(t *testing.T) {
	actual, err := clconf.ReadEnvVars()
	if err != nil {
		t.Fatal(err)
	}
	if len(actual) > 0 {
		t.Errorf("ReadEnvVars empty failed")
	}
}

func TestReadEnvVarsDoesNotExist(t *testing.T) {
	_, err := clconf.ReadEnvVars("NOT_AN_ENV_VAR_OR_PROBABLY_SHOULDNT_BE")
	if err == nil {
		t.Errorf("ReadEnvVars does not exist should have paniced")
	}
}

func TestReadEnvVarsTempValues(t *testing.T) {
	names := []string{"FOO", "BAZ"}
	values := []string{"bar", "qux"}
	defer func() {
		for _, name := range names {
			os.Unsetenv(name)
		}
	}()

	for index, name := range names {
		os.Setenv(name, values[index])
	}
	actual, err := clconf.ReadEnvVars(names...)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(values, actual) {
		t.Errorf("ReadEnvVars FOO BAZ failed: [%v] [%v]", values, actual)
	}
}

func TestReadFiles(t *testing.T) {
	actual, err := clconf.ReadFiles()
	if err != nil || len(actual) > 0 {
		t.Errorf("ReadFiles empty failed")
	}
	if _, err := clconf.ReadFiles("NOT_A_FILE_OR_PROBABLY_SHOULDNT_BE"); err == nil {
		t.Errorf("ReadFiles does not exist should have paniced")
	}
}

func TestReadFilesTempValues(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "clconf")
	if err != nil {
		t.Errorf("Unable to create temp dir: %v", err)
	}
	defer func() {
		os.RemoveAll(tempDir)
	}()

	names := []string{path.Join(tempDir, "foo"), path.Join(tempDir, "baz")}
	values := []string{"bar", "qux"}
	for index, name := range names {
		err := ioutil.WriteFile(name, []byte(values[index]), 0700)
		if err != nil {
			t.Errorf("unable to write %s: %v", name, err)
		}
	}
	actual, err := clconf.ReadFiles(names...)
	if err != nil || !reflect.DeepEqual(values, actual) {
		t.Errorf("ReadFiles foo baz failed: [%v] [%v]", values, actual)
	}
}

func TestSaveConf(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "clconf")
	if err != nil {
		t.Errorf("Unable to create temp dir: %v", err)
	}
	defer func() {
		os.RemoveAll(tempDir)
	}()

	config := map[interface{}]interface{}{"a": "b"}
	file := filepath.Join(tempDir, "config.yml")
	err = clconf.SaveConf(config, file)
	if err != nil {
		t.Errorf("SafeConf failed: %v", err)
	}
	actual, err := ioutil.ReadFile(file)
	if err != nil {
		t.Errorf("SafeConf failed, unable to read %v", file)
	}
	expected := "a: b\n"
	if expected != string(actual) {
		t.Errorf("SafeConf failed, unexpected config: %v", string(actual))
	}
}

func TestSetValue(t *testing.T) {
	expected := map[interface{}]interface{}{"foo": "baz"}
	actual := map[interface{}]interface{}{}
	err := clconf.SetValue(actual, "", map[interface{}]interface{}{"foo": "baz"})
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("SetValue empty config no path should replace root failed [%v] != [%v]: %v", expected, actual, err)
	}

	expected = map[interface{}]interface{}{
		"foo": map[interface{}]interface{}{"bar": "baz"}}
	actual = map[interface{}]interface{}{}
	err = clconf.SetValue(actual, "/foo/bar", "baz")
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("SetValue empty config failed [%v] != [%v]: %v", expected, actual, err)
	}

	actual = map[interface{}]interface{}{"foo": "bar"}
	err = clconf.SetValue(actual, "/foo/bar", "baz")
	if err == nil {
		t.Error("SetValue non map parent should have failed")
	}

	expected = map[interface{}]interface{}{"foo": "baz"}
	actual = map[interface{}]interface{}{"foo": "bar"}
	err = clconf.SetValue(actual, "/foo", "baz")
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("SetValue replace value [%v] != [%v]: %v", expected, actual, err)
	}

	expected = map[interface{}]interface{}{"foo": "bar", "hip": "hop"}
	actual = map[interface{}]interface{}{"foo": "bar"}
	err = clconf.SetValue(actual, "/hip", "hop")
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("SetValue add value [%v] != [%v]: %v", expected, actual, err)
	}
}

func testToKvMap(t *testing.T, input, expected interface{}, message string) {
	actual := clconf.ToKvMap(input)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("ToKvMap %s failed: [%v] != [%v]", message, expected, actual)
	}
}

func TestToKvMap(t *testing.T) {
	testToKvMap(t, nil, map[string]string{"/": ""}, "nil")
	testToKvMap(t, "foo", map[string]string{"/": "foo"}, "string")
	testToKvMap(t, 2, map[string]string{"/": "2"}, "number")
	testToKvMap(t,
		map[interface{}]interface{}{
			"a": "b",
			"c": 2,
		},
		map[string]string{
			"/a": "b",
			"/c": "2",
		}, "simple map")
	testToKvMap(t,
		map[interface{}]interface{}{
			"a": "b",
			"c": 2,
			"d": map[interface{}]interface{}{
				"e": "f",
				"g": 2,
			},
		},
		map[string]string{
			"/a":   "b",
			"/c":   "2",
			"/d/e": "f",
			"/d/g": "2",
		}, "multi-level map")
	testToKvMap(t,
		map[interface{}]interface{}{
			"a": "b",
			"c": 2,
			"d": map[interface{}]interface{}{
				"e": []interface{}{"f", 2, 2.2},
			},
		},
		map[string]string{
			"/a":     "b",
			"/c":     "2",
			"/d/e/0": "f",
			"/d/e/1": "2",
			"/d/e/2": "2.2",
		}, "multi-level map with array")
	testToKvMap(t,
		map[interface{}]interface{}{
			1: "one",
			2: "two",
		},
		map[string]string{
			"/1": "one",
			"/2": "two",
		}, "numeric keys")
}

func TestUnmarshalSingleYaml(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		yamlObj, err := clconf.UnmarshalSingleYaml("---")
		if err != nil {
			t.Errorf("failed to unmarshal; %v", err)
		}
		if yamlObj != nil {
			t.Errorf("expected <nil> for empty `---`, got %T: %v", yamlObj, yamlObj)
		}
	})
	t.Run("string", func(t *testing.T) {
		actual, err := clconf.UnmarshalSingleYaml("foo")
		if err != nil {
			t.Errorf("failed to unmarshal; %v", err)
		}
		expected := "foo"
		if expected != actual {
			t.Errorf("%v != %v", expected, actual)
		}
	})
	t.Run("number", func(t *testing.T) {
		actual, err := clconf.UnmarshalSingleYaml("10")
		if err != nil {
			t.Errorf("failed to unmarshal; %v", err)
		}
		expected := 10
		if expected != actual {
			t.Errorf("%v != %v", expected, actual)
		}
	})
	t.Run("array", func(t *testing.T) {
		yamlObj, err := clconf.UnmarshalSingleYaml("[\"bar\", \"baz\"]")
		if err != nil {
			t.Errorf("failed to unmarshal; %v", err)
		}
		if actual, ok := yamlObj.([]interface{}); !ok {
			t.Errorf("%v not an slice", yamlObj)
		} else {
			expected := []interface{}{"bar", "baz"}
			if !reflect.DeepEqual(expected, actual) {
				t.Errorf("%v != %v", expected, actual)
			}
		}
	})
}

func TestUnmarshalYaml(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		expected := map[interface{}]interface{}{}
		actual, err := clconf.UnmarshalYaml("---")
		if err != nil {
			t.Fatalf("Error UnmarshalYaml: %v", err)
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Empty failed: [%#v] != [%#v]", expected, actual)
		}
	})
	t.Run("configMapAndSecrets", func(t *testing.T) {
		actual, err := clconf.UnmarshalYaml(configMap, secrets)
		if err != nil || !reflect.DeepEqual(actual, configMapAndSecretsExpected) {
			t.Errorf("ConfigMap and Secrets failed: [%v] != [%v]", configMapAndSecretsExpected, actual)
		}
	})
}

func TestUnmarshalYamlInterface(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		expected := map[interface{}]interface{}{}
		actual, err := clconf.UnmarshalYamlInterface("---")
		if err != nil {
			t.Fatalf("Error UnmarshalYaml: %v", err)
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Empty failed: [%#v] != [%#v]", expected, actual)
		}
	})
	t.Run("configMapAndSecrets", func(t *testing.T) {
		actual, err := clconf.UnmarshalYamlInterface(configMap, secrets)
		if err != nil || !reflect.DeepEqual(actual, configMapAndSecretsExpected) {
			t.Errorf("ConfigMap and Secrets failed: [%#v] != [%#v]", configMapAndSecretsExpected, actual)
		}
	})
}

func TestUnmarshalYamlMultipleDocs(t *testing.T) {
	expected := map[interface{}]interface{}{
		"a": "foo",
		"b": "bar",
		"c": "stuff",
	}

	merged, err := clconf.UnmarshalYaml("---\na: bar\n---\nb: bar", "---\na: foo\n---\nc: stuff\n")
	if err != nil || !reflect.DeepEqual(merged, expected) {
		t.Errorf("Multiple docs failed: [%v] != [%v]", expected, merged)
	}
}

func TestUnmarshalYamlInterfaceMultipleDocs(t *testing.T) {
	t.Run("hashes", func(t *testing.T) {
		expected := map[interface{}]interface{}{
			"a": "foo",
			"b": "bar",
			"c": "stuff",
		}
		merged, err := clconf.UnmarshalYamlInterface("---\na: bar\n---\nb: bar", "---\na: foo\n---\nc: stuff\n")
		if err != nil || !reflect.DeepEqual(merged, expected) {
			t.Errorf("Multiple docs failed: [%#v] != [%#v]", expected, merged)
		}
	})
	t.Run("hashes then array", func(t *testing.T) {
		expected := []interface{}{
			"one",
			"two",
		}
		merged, err := clconf.UnmarshalYamlInterface("---\na: bar\n---\nb: bar", "---\na: foo\n---\nc: stuff\n---\n- one\n- two")
		if err != nil || !reflect.DeepEqual(merged, expected) {
			t.Errorf("Multiple docs failed: [%#v] != [%#v]", expected, merged)
		}
	})
	t.Run("hashes then nil", func(t *testing.T) {
		expected := map[interface{}]interface{}{
			"a": "foo",
			"b": "bar",
			"c": "stuff",
		}
		merged, err := clconf.UnmarshalYamlInterface("---\na: bar\n---\nb: bar", "---\na: foo\n---\nc: stuff\n---")
		if err != nil || !reflect.DeepEqual(merged, expected) {
			t.Errorf("Multiple docs failed: [%#v] != [%#v]", expected, merged)
		}
	})
}

func TestUnmarshalYamlJson(t *testing.T) {
	expected := map[interface{}]interface{}{
		"a": "bar",
	}

	merged, err := clconf.UnmarshalYaml(`{"a": "bar",}`)
	if err != nil || !reflect.DeepEqual(merged, expected) {
		t.Errorf("Json failed: [%v] != [%v]: %v", expected, merged, err)
	}
}

func TestUnmarshalYamlMultipleJsons(t *testing.T) {
	expected := map[interface{}]interface{}{
		"a": "bar",
		"b": "foo",
	}

	merged, err := clconf.UnmarshalYaml("{\"a\": \"bar\",}\n---\n{\"b\": \"foo\"}")
	if err != nil || !reflect.DeepEqual(merged, expected) {
		t.Errorf("Multiple json failed: [%v] != [%v]: %v", expected, merged, err)
	}
}

func TestUnmarshalYamlNumericKey(t *testing.T) {
	expected := map[interface{}]interface{}{
		1: "one",
		2: "two",
	}
	actual, err := clconf.UnmarshalYaml("1: one\n2: two")
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("Numeric keys failed: [%v] != [%v]", expected, actual)
	}
}
