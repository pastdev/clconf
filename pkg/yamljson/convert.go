// Package json provides json specific functionality like rfc 6902 patching and
// convertion between json and yaml formatted maps.
package yamljson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"unicode"

	"gopkg.in/yaml.v2"
)

var (
	JSONObjectPrefix = []byte("{")
	JSONArrayPrefix  = []byte("[")
)

// YAMLToJSON converts yaml bytes to json bytes. This is useful for other
// packages that have json specific unmarshaling code (ie: json-patch). The
// approach is inspired by:
//   https://github.com/kubernetes-sigs/yaml/blob/9535b3b1e2893fe44efb37c5c9f5665e245d786a/yaml.go
func YAMLToJSON(data []byte) ([]byte, error) {
	trimmed := bytes.TrimLeftFunc(data, unicode.IsSpace)
	if bytes.HasPrefix(trimmed, JSONArrayPrefix) || bytes.HasPrefix(trimmed, JSONObjectPrefix) {
		return data, nil
	}

	var yamlObj interface{}
	err := yaml.Unmarshal(data, &yamlObj)
	if err != nil {
		return data, fmt.Errorf("yaml unmarshal: %w", err)
	}

	v, err := json.Marshal(ConvertMapIToMapS(yamlObj))
	if err != nil {
		return data, fmt.Errorf("json marshal: %w", err)
	}
	return v, nil
}

// ConvertMapIToMapS will convert a map of the format created by yaml.Unmarshal
// to the format created by json.Unmarshal. Specifically, json uses string keys,
// while yaml uses interface{} keys.
//   https://stackoverflow.com/a/40737676/516433
func ConvertMapIToMapS(mapI interface{}) interface{} {
	switch x := mapI.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = ConvertMapIToMapS(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = ConvertMapIToMapS(v)
		}
	}
	return mapI
}

// ConvertMapSToMapI will convert a map of the format created by json.Unmarshal
// to the format created by yaml.Unmarshal. Specifically, json uses string keys,
// while yaml uses interface{} keys.
//   https://stackoverflow.com/a/40737676/516433
func ConvertMapSToMapI(mapS interface{}) interface{} {
	switch x := mapS.(type) {
	case map[string]interface{}:
		m2 := map[interface{}]interface{}{}
		for k, v := range x {
			m2[k] = ConvertMapSToMapI(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = ConvertMapSToMapI(v)
		}
	}
	return mapS
}
