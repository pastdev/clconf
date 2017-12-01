package clconf_test

import (
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

func TestMarshalYaml(t *testing.T) {
	value := map[interface{}]interface{}{"a": "b"}
	yaml, err := clconf.MarshalYaml(value)
	if err != nil || yaml != "a: b\n" {
		t.Errorf("Marshal failed for [%v]: [%v] [%v]", value, err, yaml)
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
	ok = clconf.FillValue("z", conf, &zee)
	if ok {
		t.Error("FillValue invalid path should have failed")
	}

	ok = clconf.FillValue("a", conf, &zee)
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

	value, ok = clconf.GetValue("a", conf)
	if !ok || value != "Yup" {
		t.Errorf("GetValue first level string failed: [%v] [%v]", ok, value)
	}

	value, ok = clconf.GetValue("a/f", conf)
	if ok {
		t.Errorf("GetValue non map indexing should have failed: [%v] [%v]", ok, value)
	}

	value, ok = clconf.GetValue("z", conf)
	if ok {
		t.Errorf("GetValue missing have failed: [%v] [%v]", ok, value)
	}
}
