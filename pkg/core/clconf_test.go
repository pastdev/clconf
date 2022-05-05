package core_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/imdario/mergo"
	"github.com/mitchellh/mapstructure"
	"github.com/pastdev/clconf/v2/pkg/core"
	"github.com/pastdev/clconf/v2/pkg/yamljson"
)

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

func assertMergeValue(
	t *testing.T,
	expected interface{},
	conf interface{},
	path string,
	value interface{},
	overwrite bool,
) {
	expectedYaml, _ := yamljson.MarshalYaml(expected)
	confYaml, _ := yamljson.MarshalYaml(conf)
	valueYaml, _ := yamljson.MarshalYaml(value)
	err := core.MergeValue(conf, path, value, overwrite)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if !reflect.DeepEqual(expected, conf) {
		actualYaml, _ := yamljson.MarshalYaml(conf)
		t.Errorf(
			"\nexpected:\n---\n%v\nactual:\n---\n%v\nconf:\n---\n%v\nvalue at %s\n---\n%v",
			string(expectedYaml),
			string(actualYaml),
			string(confYaml),
			path,
			string(valueYaml))
	}
}

func TestFill(t *testing.T) {
	expectedTimestampString := "2019-10-16T12:28:49Z"
	conf, err := yamljson.UnmarshalYamlInterface("---\n" +
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
	err = core.Fill(
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
	conf, _ := yamljson.UnmarshalYamlInterface(yaml1and2)

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
	ok := core.FillValue("", conf, &root)
	expectedRoot := Root{A: "Yup", B: Bee{C: 2, E: 2, F: Eff{G: "foobar"}}}
	if !ok || !reflect.DeepEqual(expectedRoot, root) {
		t.Errorf("FillValue empty path failed: [%v] [%v] == [%v]", ok, expectedRoot, root)
	}

	var bee Bee
	ok = core.FillValue("b", conf, &bee)
	expectedBee := Bee{C: 2, E: 2, F: Eff{G: "foobar"}}
	if !ok || !reflect.DeepEqual(expectedBee, bee) {
		t.Errorf("FillValue first level failed: [%v] [%v] == [%v]", ok, expectedBee, bee)
	}

	type BeeLight struct {
		C int
		E int
	}
	var beeLight BeeLight
	ok = core.FillValue("b", conf, &beeLight)
	expectedBeeLight := BeeLight{C: 2, E: 2}
	if !ok || !reflect.DeepEqual(expectedBeeLight, beeLight) {
		t.Errorf("FillValue first level string failed: [%v] [%v] == [%v]", ok, expectedBeeLight, beeLight)
	}

	type Zee struct {
		ShouldntWork string
	}
	var zee Zee
	ok = core.FillValue("/z", conf, &zee)
	if ok {
		t.Error("FillValue invalid path should have failed")
	}

	ok = core.FillValue("a/", conf, &zee)
	if ok {
		t.Error("FillValue a should not have been z")
	}
}

func TestGetValue(t *testing.T) {
	conf, _ := yamljson.UnmarshalYamlInterface(yaml1and2)

	value, err := core.GetValue(conf, "")
	if err != nil || !reflect.DeepEqual(conf, value) {
		t.Errorf("GetValue empty path failed: [%v] [%v] == [%v]", err, conf, value)
	}

	value, err = core.GetValue(conf, "/")
	if err != nil || !reflect.DeepEqual(conf, value) {
		t.Errorf("GetValue empty path failed: [%v] [%v] == [%v]", err, conf, value)
	}

	value, err = core.GetValue(conf, "/a")
	if err != nil || value != "Yup" {
		t.Errorf("GetValue first level string failed: [%v] [%v]", err, value)
	}

	value, err = core.GetValue(conf, "/b//f//g")
	if err != nil || value != "foobar" {
		t.Errorf("GetValue third level string multi slash failed: [%v] [%v]", err, value)
	}

	value, err = core.GetValue(conf, "/a/f")
	if err == nil {
		t.Errorf("GetValue non map indexing should have failed: [%v] [%v]", err, value)
	}

	value, err = core.GetValue(conf, "/z")
	if err == nil {
		t.Errorf("GetValue missing have failed: [%v] [%v]", err, value)
	}

	conf, _ = yamljson.UnmarshalYamlInterface(yamlWithList)

	value, err = core.GetValue(conf, "/a")
	expected := []interface{}{"one", "two", "three"}
	if err != nil || !reflect.DeepEqual(expected, value) {
		t.Errorf("GetValue list failed: (err:[%v]) [%v] == [%v]", err, expected, value)
	}

	value, err = core.GetValue(conf, "/a/0")
	stringExpected := "one"
	if err != nil || !reflect.DeepEqual(stringExpected, value) {
		t.Errorf("GetValue list item failed: (err:[%v]) [%v] == [%v]", err, stringExpected, value)
	}

	_, err = core.GetValue(conf, "/a/b")
	if err == nil {
		t.Errorf("GetValue list item invalid index should have failed")
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

func TestSaveConf(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "core")
	if err != nil {
		t.Errorf("Unable to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	config := map[interface{}]interface{}{"a": "b"}
	file := filepath.Join(tempDir, "config.yml")
	err = core.SaveConf(config, file)
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
	err := core.SetValue(actual, "", map[interface{}]interface{}{"foo": "baz"})
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("SetValue empty config no path should replace root failed [%v] != [%v]: %v", expected, actual, err)
	}

	expected = map[interface{}]interface{}{
		"foo": map[interface{}]interface{}{"bar": "baz"}}
	actual = map[interface{}]interface{}{}
	err = core.SetValue(actual, "/foo/bar", "baz")
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("SetValue empty config failed [%v] != [%v]: %v", expected, actual, err)
	}

	actual = map[interface{}]interface{}{"foo": "bar"}
	err = core.SetValue(actual, "/foo/bar", "baz")
	if err == nil {
		t.Error("SetValue non map parent should have failed")
	}

	expected = map[interface{}]interface{}{"foo": "baz"}
	actual = map[interface{}]interface{}{"foo": "bar"}
	err = core.SetValue(actual, "/foo", "baz")
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("SetValue replace value [%v] != [%v]: %v", expected, actual, err)
	}

	expected = map[interface{}]interface{}{"foo": "bar", "hip": "hop"}
	actual = map[interface{}]interface{}{"foo": "bar"}
	err = core.SetValue(actual, "/hip", "hop")
	if err != nil || !reflect.DeepEqual(expected, actual) {
		t.Errorf("SetValue add value [%v] != [%v]: %v", expected, actual, err)
	}
}

func testToKvMap(t *testing.T, input, expected interface{}, message string) {
	actual := core.ToKvMap(input)
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
