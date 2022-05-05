package yamljson_test

import (
	"testing"

	"github.com/pastdev/clconf/v3/pkg/yamljson"
)

func TestMarshalYaml(t *testing.T) {
	value := map[interface{}]interface{}{"a": "b"}
	yaml, err := yamljson.MarshalYaml(value)
	if err != nil || string(yaml) != "a: b\n" {
		t.Errorf("Marshal failed for [%v]: [%v] [%v]", value, err, yaml)
	}
}
