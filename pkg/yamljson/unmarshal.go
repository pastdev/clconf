package yamljson

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v2"
)

// UnmarshalSingleYaml will unmarshal the first yaml doc in a single yaml/json
// string without merging. This form works for any yaml data, not just objects.
func UnmarshalSingleYaml(yamlString string) (interface{}, error) {
	results, err := UnmarshalAllYaml(yamlString)
	return results[0], err
}

// UnmarshalAllYaml will unmarshal all yaml docs in a single yaml/json
// string without merging. This form works for any yaml data, not just objects.
func UnmarshalAllYaml(yamlString string) ([]interface{}, error) {
	var results []interface{}
	var err error
	decoder := yaml.NewDecoder(strings.NewReader(yamlString))
	for err == nil {
		var result interface{}
		err = decoder.Decode(&result)
		if err == nil {
			results = append(results, result)
		}
	}

	if errors.Is(err, io.EOF) {
		return results, nil
	}
	if err != nil {
		return results, fmt.Errorf("yaml decode: %w", err)
	}
	return results, nil
}

// UnmarshalYamlInterface will parse all the supplied yaml strings, merge the resulting
// objects, and return the resulting map. If a root node is a list it will be
// converted to an int map prior to merging. An emtpy document returns nil.
func UnmarshalYamlInterface(yamlStrings ...string) (interface{}, error) {
	// We collect all the yamls into a base string map so mergo can handle them as
	// subnodes for consistentcy (mergo doesn't like conflicting types in root
	// nodes)
	var allYamls []map[string]interface{}
	for _, yamlString := range yamlStrings {
		yamls, err := UnmarshalAllYaml(yamlString)
		if err != nil {
			return nil, err
		}
		for _, yaml := range yamls {
			// We do this to maintain backward compatibility with empty docs being
			// treated as an empty map
			if yaml != nil {
				allYamls = append(allYamls, map[string]interface{}{"root": yaml})
			}
		}
	}

	result := make(map[string]interface{})
	for _, y := range allYamls {
		if err := mergo.Merge(&result, y, mergo.WithOverride); err != nil {
			return nil, fmt.Errorf("yaml merge: %w", err)
		}
	}
	r := result["root"]
	if r == nil {
		// We do this to maintain backward compatibility with empty docs being
		// treated as an empty map
		return map[interface{}]interface{}{}, nil
	}
	return r, nil
}
