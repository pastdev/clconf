// Package core provides functions to extract values from a set of yaml
// files after merging them.
package core

import (
	"fmt"
	"math"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"

	"dario.cat/mergo"
	"github.com/mitchellh/mapstructure"
	"github.com/pastdev/clconf/v3/pkg/yamljson"
)

// Fill will fill a according to DecoderConfig with the values from conf.
func Fill(keyPath string, conf interface{}, decoderConfig *mapstructure.DecoderConfig) error {
	value, err := GetValue(conf, keyPath)
	if err != nil {
		return err
	}

	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return fmt.Errorf("create decoder: %w", err)
	}

	err = decoder.Decode(value)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	return nil
}

// FillValue will fill a struct, out, with values from conf.
func FillValue(keyPath string, conf interface{}, out interface{}) bool {
	err := Fill(keyPath, conf, &mapstructure.DecoderConfig{Result: out})
	return err == nil
}

// GetValue returns the value at the indicated path.  Paths are separated by
// the '/' character.  The empty string or "/" will return conf itself.
func GetValue(conf interface{}, keyPath string) (interface{}, error) {
	if keyPath == "" || keyPath == "/" {
		return conf, nil
	}

	var value = conf
	currentPath := "/"
	for _, part := range strings.Split(keyPath, "/") {
		if part == "" {
			continue
		}

		currentPath = path.Join(currentPath, part)

		switch typed := value.(type) {
		case map[string]interface{}:
			// json deserialized
			var ok bool
			value, ok = typed[part]
			if !ok {
				return nil, fmt.Errorf(
					"value at [%v] does not exist",
					currentPath)
			}
		case map[interface{}]interface{}:
			// yaml deserialized
			var ok bool
			value, ok = typed[part]
			if ok {
				continue
			}
			intKey, err := strconv.ParseInt(part, 10, 64)
			if err == nil {
				if intKey >= math.MinInt && intKey <= math.MaxInt {
					value, ok = typed[int(intKey)]
					if ok {
						continue
					}
				}
				if intKey >= math.MinInt8 && intKey <= math.MaxInt8 {
					value, ok = typed[int8(intKey)]
					if ok {
						continue
					}
				}
				if intKey >= math.MinInt16 && intKey <= math.MaxInt16 {
					value, ok = typed[int16(intKey)]
					if ok {
						continue
					}
				}
				if intKey >= math.MinInt32 && intKey <= math.MaxInt32 {
					value, ok = typed[int32(intKey)]
					if ok {
						continue
					}
				}
				value, ok = typed[int64(intKey)]
				if ok {
					continue
				}
			}
			boolKey, err := strconv.ParseBool(part)
			if err == nil {
				value, ok = typed[boolKey]
				if ok {
					continue
				}
			}
			return nil, fmt.Errorf(
				"value at [%v] does not exist",
				currentPath)
		case []interface{}:
			i, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf(
					"value at [%v] is array, but index [%v] is not int: %w",
					path.Dir(currentPath),
					part,
					err)
			}
			value = typed[i]
		default:
			return nil, fmt.Errorf(
				"value at [%v] not a map or slice: %v",
				part,
				reflect.ValueOf(value).Kind())
		}
	}
	return value, nil
}

// ListToMap converts a list to an integer map.
func ListToMap(l []interface{}) map[interface{}]interface{} {
	m := make(map[interface{}]interface{})
	for i, v := range l {
		m[i] = v
	}
	return m
}

func splitKeyPath(keyPath string) ([]string, string) {
	if keyPath == "" || keyPath == "/" {
		return []string{}, ""
	}

	parts := []string{}

	for _, parentPart := range strings.Split(keyPath, "/") {
		if parentPart == "" {
			continue
		}
		parts = append(parts, parentPart)
	}

	lastIndex := len(parts) - 1
	if lastIndex >= 0 {
		return parts[:lastIndex], parts[lastIndex]
	}
	return parts, keyPath
}

// MergeValue will merge the values from value into config at keyPath.
// If overwrite is true, values from value will overwrite existing
// values in config.
func MergeValue(config interface{}, keyPath string, value interface{}, overwrite bool) error {
	parent, key, err := getParentAndKey(config, keyPath)
	if err != nil {
		return err
	}

	// ensure the root element is always a map for the merge library
	if key == "" {
		parent = map[interface{}]interface{}{key: parent}
		value = map[interface{}]interface{}{key: value}
	} else {
		parent = map[interface{}]interface{}{key: parent[key]}
		value = map[interface{}]interface{}{key: value}
	}

	err = mergo.Merge(
		&parent,
		value,
		func(c *mergo.Config) { c.Overwrite = overwrite })
	if err != nil {
		return fmt.Errorf("merge: %w", err)
	}
	return nil
}

func getParentAndKey(config interface{}, keyPath string) (map[interface{}]interface{}, string, error) {
	configMap, ok := config.(map[interface{}]interface{})
	if !ok {
		return nil, "", fmt.Errorf("config not a map")
	}
	parentParts, key := splitKeyPath(keyPath)
	if key == "" {
		return configMap, "", nil
	}

	parent := configMap
	for i, parentPart := range parentParts {
		parentValue, ok := parent[parentPart]
		if !ok || parentValue == nil {
			parentValue = make(map[interface{}]interface{})
			parent[parentPart] = parentValue
		}
		valueMap, ok := parentValue.(map[interface{}]interface{})
		if !ok {
			return nil, "", fmt.Errorf(
				"parent at /%s not a map (type: %T)",
				strings.Join(parentParts[0:i+1], "/"),
				parentValue)
		}

		parent = valueMap
	}

	return parent, key, nil
}

// SetValue will set the value of config at keyPath to value.
func SetValue(config interface{}, keyPath string, value interface{}) error {
	parent, key, err := getParentAndKey(config, keyPath)
	if err != nil {
		return err
	}
	if key == "" {
		valueMap, ok := value.(map[interface{}]interface{})
		if !ok {
			return fmt.Errorf("if replacing root, value must be a map")
		}
		for k := range parent {
			delete(parent, k)
		}
		for k := range valueMap {
			parent[k] = valueMap[k]
		}
		return nil
	}

	parent[key] = value

	return nil
}

// SaveConf will save config to file as yaml
func SaveConf(config interface{}, file string) error {
	yamlBytes, err := yamljson.MarshalYaml(config)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	err = os.WriteFile(file, yamlBytes, 0660) //nolint:gosec
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

// ToKvMap will return a one-level map of key value pairs where the key is
// a / separated path of subkeys.
func ToKvMap(conf interface{}) map[string]string {
	kvMap := make(map[string]string)
	Walk(func(keyStack []string, value interface{}) {
		key := "/" + strings.Join(keyStack, "/")
		if value == nil {
			kvMap[key] = ""
		} else {
			kvMap[key] = fmt.Sprintf("%v", value)
		}
	}, conf)
	return kvMap
}

// Walk will recursively iterate over all the nodes of conf calling callback
// for each node.
func Walk(callback func(key []string, value interface{}), conf interface{}) {
	walk(callback, conf, []string{})
}

func walk(callback func(key []string, value interface{}), node interface{}, keyStack []string) {
	switch typed := node.(type) {
	case map[string]interface{}:
		// json deserialized
		for k, v := range typed {
			keyStack := append(keyStack, fmt.Sprintf("%v", k))
			walk(callback, v, keyStack)
		}
	case map[interface{}]interface{}:
		// yaml deserialized
		for k, v := range typed {
			keyStack := append(keyStack, fmt.Sprintf("%v", k))
			walk(callback, v, keyStack)
		}
	case []interface{}:
		for i, j := range typed {
			keyStack := append(keyStack, strconv.Itoa(i))
			walk(callback, j, keyStack)
		}
	default:
		callback(keyStack, node)
	}
}
