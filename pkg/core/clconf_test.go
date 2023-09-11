package core_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"dario.cat/mergo"
	"github.com/mitchellh/mapstructure"
	"github.com/pastdev/clconf/v3/pkg/core"
	"github.com/pastdev/clconf/v3/pkg/yamljson"
	"github.com/stretchr/testify/require"
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
const yamlWithNonStringKeys = "" +
	"a:\n" +
	"  1234:\n" +
	"    foo: bar\n" +
	"  5578: 91011\n" +
	"  true: really?"

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
	test := func(name string, conf interface{}, path string, expected interface{}, errExpected bool) {
		t.Run(name, func(t *testing.T) {
			actual, err := core.GetValue(conf, path)
			if errExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, expected, actual)
		})
	}

	conf, err := yamljson.UnmarshalYamlInterface(yaml1and2)
	require.NoError(t, err, "parsing yaml1and2")

	test("empty path", conf, "", conf, false)
	test("slash path", conf, "/", conf, false)
	test("first level string", conf, "/a", "Yup", false)
	test("third level string multi slash", conf, "/b//f//g", "foobar", false)
	test("non map indexing", conf, "/a/f", "", true)
	test("missing", conf, "/z", "", true)

	conf, err = yamljson.UnmarshalYamlInterface(yamlWithList)
	require.NoError(t, err, "parsing yamlWithList")

	test("list", conf, "/a", []interface{}{"one", "two", "three"}, false)
	test("list item", conf, "/a/0", "one", false)
	test("list item invalid index", conf, "/a/b", "", true)

	conf, err = yamljson.UnmarshalYamlInterface(yamlWithNonStringKeys)
	require.NoError(t, err, "parsing yaml with int keys")

	test("int key with nested", conf, "/a/1234/foo", "bar", false)
	test("int key", conf, "/a/5578", 91011, false)
	test("bool key", conf, "/a/true", "really?", false)
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
