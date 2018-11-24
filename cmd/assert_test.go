package cmd

import (
	"reflect"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func assertYamlEqual(t *testing.T, message, expected, actual string) {
	var expectedUnmarshaled interface{}
	err := yaml.Unmarshal([]byte(expected), &expectedUnmarshaled)
	if err != nil {
		t.Errorf("%s unable to unmarshal expected %s: %s", message, err, expected)
		return
	}
	var actualUnmarshaled interface{}
	err = yaml.Unmarshal([]byte(actual), &actualUnmarshaled)
	if err != nil {
		t.Errorf("%s unable to unmarshal actual %s: %s", message, err, actual)
		return
	}

	if !reflect.DeepEqual(expectedUnmarshaled, actualUnmarshaled) {
		t.Errorf("%s yaml not equivalent: %v != %v", message, expectedUnmarshaled, actualUnmarshaled)
	}
}
