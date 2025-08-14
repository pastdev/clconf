package yamljson

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v2"
)

// MarshalYaml will convert an object to yaml
func MarshalYaml(in interface{}) ([]byte, error) {
	var v bytes.Buffer
	enc := yaml.NewEncoder(&v)
	err := enc.Encode(in)
	if err != nil {
		return v.Bytes(), fmt.Errorf("yaml marshal: %w", err)
	}
	return v.Bytes(), nil
}
