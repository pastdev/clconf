package yamljson

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// MarshalYaml will convert an object to yaml
func MarshalYaml(in interface{}) ([]byte, error) {
	v, err := yaml.Marshal(in)
	if err != nil {
		return v, fmt.Errorf("yaml marshal: %w", err)
	}
	return v, nil
}
