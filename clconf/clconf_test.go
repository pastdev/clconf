package clconf_test

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"

	"gitlab.com/pastdev/s2i/clconf/clconf"
)

const yaml1 = "" +
	"a: Nope\n" +
	"b:\n" +
	"  c: 2\n"
const yaml2 = "" +
	"a: Yup\n" +
	"b:\n" +
	"  e: 2\n" +
	"  f:\n" +
	"    g: foobar\n"
const yaml1and2 = "" +
	"a: Yup\n" +
	"b:\n" +
	"  c: 2\n" +
	"  e: 2\n" +
	"  f:\n" +
	"    g: foobar\n"
const yaml2and1 = "" +
	"a: Nope\n" +
	"b:\n" +
	"  c: 2\n" +
	"  e: 2\n" +
	"  f:\n" +
	"    g: foobar\n"

func TestBase64Strings(t *testing.T) {
	expected := []string{}
	encoded := []string{}
	actual := clconf.DecodeBase64Strings(encoded...)
	if len(actual) != 0 {
		t.Errorf("Base64Strings empty failed: [%v]", actual)
	}

	expected = []string{"one", "two"}
	encoded = []string{
		base64.StdEncoding.EncodeToString([]byte(expected[0])),
		base64.StdEncoding.EncodeToString([]byte(expected[1]))}
	actual = clconf.DecodeBase64Strings(encoded...)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Base64Strings one two failed: [%v] == [%v]", expected, actual)
	}
}

func TestBase64StringsPanic(t *testing.T) {
	defer func() {
		recover()
	}()
	clconf.DecodeBase64Strings("&*INVALID*&")
	t.Error("Base64Strings invalid did not panic")
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

	value, ok := clconf.GetValue("", conf)
	if !ok || !reflect.DeepEqual(conf, value) {
		t.Errorf("GetValue empty path failed: [%v] [%v] == [%v]", ok, conf, value)
	}

	value, ok = clconf.GetValue("/", conf)
	if !ok || !reflect.DeepEqual(conf, value) {
		t.Errorf("GetValue empty path failed: [%v] [%v] == [%v]", ok, conf, value)
	}

	value, ok = clconf.GetValue("/a", conf)
	if !ok || value != "Yup" {
		t.Errorf("GetValue first level string failed: [%v] [%v]", ok, value)
	}

	value, ok = clconf.GetValue("/b//f//g", conf)
	if !ok || value != "foobar" {
		t.Errorf("GetValue third level string multi slash failed: [%v] [%v]", ok, value)
	}

	value, ok = clconf.GetValue("/a/f", conf)
	if ok {
		t.Errorf("GetValue non map indexing should have failed: [%v] [%v]", ok, value)
	}

	value, ok = clconf.GetValue("/z", conf)
	if ok {
		t.Errorf("GetValue missing have failed: [%v] [%v]", ok, value)
	}
}

func TestLoadConf(t *testing.T) {
	envVars := []string{"a"}
	tempDir := os.TempDir()
	defer func() {
		os.RemoveAll(tempDir)
		for _, name := range envVars {
			os.Unsetenv(name)
		}
		os.Unsetenv("YAML_FILES")
		os.Unsetenv("YAML_VARS")
	}()

	envValues := []string{base64.StdEncoding.EncodeToString([]byte("a: env"))}
	for index, name := range envVars {
		os.Setenv(name, envValues[index])
	}

	fileVars := []string{path.Join(tempDir, "a")}
	fileValues := []string{"a: file"}
	for index, name := range fileVars {
		ioutil.WriteFile(name, []byte(fileValues[index]), 0700)
	}

	overrides := []string{base64.StdEncoding.EncodeToString([]byte("a: override"))}

	actual := clconf.LoadConf([]string{}, []string{})
	if len(actual) > 0 {
		t.Errorf("LoadConf no config failed")
	}

	expected, _ := clconf.UnmarshalYaml("a: override")
	actual = clconf.LoadConf([]string{}, overrides)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf overrides only failed: [%v] != [%v]", expected, actual)
	}

	os.Setenv("YAML_FILES", fileVars[0])
	expected, _ = clconf.UnmarshalYaml(fileValues[0])
	actual = clconf.LoadConfFromEnvironment([]string{}, []string{})
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf files only failed: [%v] != [%v]", expected, actual)
	}
	os.Unsetenv("YAML_FILES")

	os.Setenv("YAML_VARS", envVars[0])
	expected, _ = clconf.UnmarshalYaml(clconf.DecodeBase64Strings(envValues[0])...)
	actual = clconf.LoadConfFromEnvironment([]string{}, []string{})
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf env only failed: [%v] != [%v]", expected, actual)
	}
	os.Unsetenv("YAML_VARS")

	os.Setenv("YAML_FILES", fileVars[0])
	os.Setenv("YAML_VARS", envVars[0])
	expected, _ = clconf.UnmarshalYaml("a: override")
	actual = clconf.LoadConf([]string{}, overrides)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("LoadConf all failed: [%v] != [%v]", expected, actual)
	}
}

func TestMarshalYaml(t *testing.T) {
	value := map[interface{}]interface{}{"a": "b"}
	yaml, err := clconf.MarshalYaml(value)
	if err != nil || yaml != "a: b\n" {
		t.Errorf("Marshal failed for [%v]: [%v] [%v]", value, err, yaml)
	}
}

func TestReadEnvVars(t *testing.T) {
	actual := clconf.ReadEnvVars()
	if len(actual) > 0 {
		t.Errorf("ReadEnvVars empty failed")
	}
}

func TestReadEnvVarsDoesNotExist(t *testing.T) {
	defer func() {
		recover()
	}()
	clconf.ReadEnvVars("NOT_AN_ENV_VAR_OR_PROBABLY_SHOULDNT_BE")
	t.Errorf("ReadEnvVars does not exist should have paniced")
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
	actual := clconf.ReadEnvVars(names...)
	if !reflect.DeepEqual(values, actual) {
		t.Errorf("ReadEnvVars FOO BAZ failed: [%v] [%v]", values, actual)
	}
}

func TestReadFiles(t *testing.T) {
	actual := clconf.ReadFiles()
	if len(actual) > 0 {
		t.Errorf("ReadFiles empty failed")
	}
}

func TestReadFilesDoesNotExist(t *testing.T) {
	defer func() {
		recover()
	}()
	clconf.ReadFiles("NOT_A_FILE_OR_PROBABLY_SHOULDNT_BE")
	t.Errorf("ReadFiles does not exist should have paniced")
}

func TestReadFilesTempValues(t *testing.T) {
	tempDir := os.TempDir()
	defer func() {
		os.RemoveAll(tempDir)
	}()

	names := []string{path.Join(tempDir, "foo"), path.Join(tempDir, "baz")}
	values := []string{"bar", "qux"}
	for index, name := range names {
		ioutil.WriteFile(name, []byte(values[index]), 0700)
	}
	actual := clconf.ReadFiles(names...)
	if !reflect.DeepEqual(values, actual) {
		t.Errorf("ReadFiles foo baz failed: [%v] [%v]", values, actual)
	}
}

func TestUnmarshalYaml(t *testing.T) {
	_, err := clconf.UnmarshalYaml("foo")
	if err == nil {
		t.Error("Unmarshal illegal char")
	}

	expected, _ := clconf.UnmarshalYaml(yaml2and1)
	merged, err := clconf.UnmarshalYaml(yaml2, yaml1)
	if err != nil || !reflect.DeepEqual(merged, expected) {
		t.Errorf("Merge 2 and 1 failed: [%v] != [%v]", expected, merged)
	}
}
