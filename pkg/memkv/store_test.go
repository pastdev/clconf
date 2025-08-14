package memkv_test

import (
	"testing"

	"github.com/pastdev/clconf/v3/pkg/memkv"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestDel(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		s := memkv.New(memkv.WithKvMap(
			map[string]string{
				"/foo": "hop",
			}))
		assert.True(t, s.Exists("/foo"))
		s.Del("/foo")
		assert.False(t, s.Exists("/foo"))
	})
}

func TestExists(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		s := memkv.New(memkv.WithKvMap(
			map[string]string{
				"/foo": "hop",
			}))
		assert.True(t, s.Exists("/foo"))
		assert.False(t, s.Exists("/bar"))
	})
}

func TestGetAllAndGetAllValues(t *testing.T) {
	tester := func(
		name string,
		data map[string]string,
		pattern string,
		values memkv.KVPairs,
		badPattern string,
	) {
		t.Run(name, func(t *testing.T) {
			s := memkv.New(memkv.WithKvMap(data))

			v, err := s.GetAll(pattern)
			assert.NoError(t, err)
			assert.Equal(t, values, v)

			_, err = s.GetAll(badPattern)
			assert.Error(t, err)
			assert.True(t, memkv.IsBadPattern(err))

			expectedStrVals := make([]string, len(values))
			for i, kvPair := range values {
				expectedStrVals[i] = kvPair.Value
			}
			strVals, err := s.GetAllValues(pattern)
			assert.NoError(t, err)
			assert.Equal(t, expectedStrVals, strVals)

			_, err = s.GetAllValues(badPattern)
			assert.Error(t, err)
			assert.True(t, memkv.IsBadPattern(err))
		})
	}

	tester("simple",
		map[string]string{
			"/foo/bar": "hip",
			"/foo/baz": "hop",
		},
		"/foo/*",
		memkv.KVPairs{
			{Key: "/foo/bar", Value: "hip"},
			{Key: "/foo/baz", Value: "hop"},
		},
		"[]bar")
}

func TestGetAndGetValue(t *testing.T) {
	tester := func(
		name string,
		data map[string]string,
		key string,
		value string,
		badKey string,
	) {
		t.Run(name, func(t *testing.T) {
			s := memkv.New(memkv.WithKvMap(data))

			v, err := s.Get(key)
			assert.NoError(t, err)
			assert.Equal(t, memkv.KVPair{Key: key, Value: value}, v)

			_, err = s.Get(badKey)
			assert.Error(t, err)
			assert.True(t, memkv.IsNotExists(err))

			strV, err := s.GetValue(key)
			assert.NoError(t, err)
			assert.Equal(t, value, strV)

			_, err = s.GetValue(badKey)
			assert.Error(t, err)
			assert.True(t, memkv.IsNotExists(err))

			strV, err = s.GetValue(badKey, "default")
			assert.NoError(t, err)
			assert.Equal(t, "default", strV)
		})
	}

	tester("simple",
		map[string]string{
			"/foo": "hop",
		},
		"/foo",
		"hop",
		"/bar")
}

func TestListAndListDir(t *testing.T) {
	tester := func(
		test string,
		data map[string]string,
		key string,
		lsExpected []string,
		lsDirExpected []string,
	) {
		t.Run(test, func(t *testing.T) {
			s := memkv.New(memkv.WithKvMap(data))
			assert.Equal(t, lsExpected, s.List(key))
			assert.Equal(t, lsDirExpected, s.ListDir(key))
		})
	}

	tester("basic",
		map[string]string{
			"/foo/bar/hip/0": "hop",
			"/foo/bar/hip/1": "hap",
			"/foobar/hip/1":  "hap",
		},
		"/",
		[]string{"foo", "foobar"},
		[]string{"foo", "foobar"})

	tester("some not dir",
		map[string]string{
			"/foo/bar/hip/0": "hop",
			"/foo/bar/hip/1": "hap",
			"/foo/hip":       "hap",
		},
		"/foo",
		[]string{"bar", "hip"},
		[]string{"bar"})

	tester("examine root",
		map[string]string{
			"/foo/bar": "hop",
			"/foo/baz": "hap",
			"/bip":     "bop",
		},
		"/",
		[]string{"bip", "foo"},
		[]string{"foo"})

	tester("prefix same but not match",
		map[string]string{
			"/foo/bar/hip/0": "hop",
			"/foobar/hip/1":  "hap",
			"/bar":           "baz",
		},
		"/foo",
		[]string{"bar"},
		[]string{"bar"})

	tester("full path",
		map[string]string{
			"/foo/bar": "hop",
		},
		"/foo/bar",
		[]string{"bar"},
		[]string{})

	tester("full path root node",
		map[string]string{
			"/foo": "hop",
		},
		"/foo",
		[]string{"foo"},
		[]string{})
}

func TestPurge(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		s := memkv.New(memkv.WithKvMap(
			map[string]string{
				"/foo": "hop",
				"/bar": "baz",
			}))
		assert.True(t, s.Exists("/foo"))
		assert.True(t, s.Exists("/bar"))
		s.Purge()
		assert.False(t, s.Exists("/foo"))
		assert.False(t, s.Exists("/bar"))
	})
}

func TestSet(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		s := memkv.New()
		assert.False(t, s.Exists("/foo"))
		s.Set("/foo", "bar")
		assert.True(t, s.Exists("/foo"))
		v, err := s.GetValue("/foo")
		assert.NoError(t, err)
		assert.Equal(t, "bar", v)
	})
}

func TestFromMapToKvMap(t *testing.T) {
	tester := func(
		test string,
		unmarshaler func(t *testing.T, v string) interface{},
		data interface{},
		expected map[string]string,
	) {
		t.Run(test, func(t *testing.T) {
			if unmarshaler != nil {
				data = unmarshaler(t, data.(string))
			}
			s := memkv.New(memkv.WithMap(data))
			assert.Equal(t, expected, s.ToKvMap())
		})
	}

	unmarshalJSON := func(t *testing.T, v string) interface{} {
		var out map[string]interface{}
		err := yaml.Unmarshal([]byte(v), &out)
		if err != nil {
			t.Fatalf("unable to json.Unmarshal: %v", err)
		}
		return out
	}

	unmarshalYaml := func(t *testing.T, v string) interface{} {
		var out map[interface{}]interface{}
		err := yaml.Unmarshal([]byte(v), &out)
		if err != nil {
			t.Fatalf("unable to yaml.Unmarshal: %v", err)
		}
		return out
	}

	tester("simple yaml deserialized",
		unmarshalYaml,
		""+
			"---\n"+
			"foo:\n"+
			"  bar: baz\n"+
			"  barint: 1\n"+
			"  barbool: true\n"+
			"  3: intkey\n",
		map[string]string{
			"/foo/bar":     "baz",
			"/foo/barint":  "1",
			"/foo/barbool": "true",
			"/foo/3":       "intkey",
		})

	tester("simple json deserialized",
		unmarshalJSON,
		"{\"foo\": {\"bar\": \"baz\", \"barint\": 1, \"barbool\": true, 3: \"intkey\"}}",
		map[string]string{
			"/foo/bar":     "baz",
			"/foo/barint":  "1",
			"/foo/barbool": "true",
			"/foo/3":       "intkey",
		})

	tester("nil value",
		nil,
		map[string]interface{}{
			"foo": nil,
		},
		map[string]string{
			"/foo": "",
		})

	tester("array",
		nil,
		map[string]interface{}{
			"foo": []interface{}{"bar", "baz"},
		},
		map[string]string{
			"/foo/0": "bar",
			"/foo/1": "baz",
		})

	tester("complex",
		nil,
		map[interface{}]interface{}{
			"foo": []interface{}{"bar", "baz"},
			"hip": map[interface{}]interface{}{
				"zip": []interface{}{"zap", 7, "zup"},
				"fiz": "fuz",
				"dip": map[interface{}]interface{}{"dup": 1},
			},
		},
		map[string]string{
			"/foo/0":       "bar",
			"/foo/1":       "baz",
			"/hip/zip/0":   "zap",
			"/hip/zip/1":   "7",
			"/hip/zip/2":   "zup",
			"/hip/fiz":     "fuz",
			"/hip/dip/dup": "1",
		})
}
