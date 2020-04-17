// Package clconf provides functions to extract values from a set of yaml
// files after merging them.
package clconf

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/imdario/mergo"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

// Splitter is the regex used to split YAML_FILES and YAML_VARS
var Splitter = regexp.MustCompile(`,`)

// ConfSources contains sources of yaml for loading. See Load() for more info
type ConfSources struct {
	// Environment loads config from environment vars
	Environment bool
	// Files to read
	Files []string
	// Overrides are Base64 encoded strings of yaml
	Overrides []string
	Stream    io.Reader
}

// Load will load the config determined by settings in the struct. In order
// of precedence (highest last), files, YAML_FILES env var, overrides,
// YAML_VARS env var, stream.
func (s ConfSources) Load() (map[interface{}]interface{}, error) {
	files := s.Files
	overrides := s.Overrides

	if s.Environment {
		if yamlFiles, ok := os.LookupEnv("YAML_FILES"); ok {
			files = append(files, Splitter.Split(yamlFiles, -1)...)
		}
		if yamlVars, ok := os.LookupEnv("YAML_VARS"); ok {
			envVars, err := ReadEnvVars(Splitter.Split(yamlVars, -1)...)
			if err != nil {
				return nil, err
			}
			overrides = append(overrides, envVars...)
		}
	}

	yamls := []string{}
	if len(files) > 0 {
		moreYamls, err := ReadFiles(files...)
		if err != nil {
			return nil, err
		}
		yamls = append(yamls, moreYamls...)
	}
	if len(overrides) > 0 {
		moreYamls, err := DecodeBase64Strings(overrides...)
		if err != nil {
			return nil, err
		}
		yamls = append(yamls, moreYamls...)
	}
	if s.Stream != nil {
		streamYaml, err := ioutil.ReadAll(s.Stream)
		if err != nil {
			return nil, fmt.Errorf("Error reading stdin: %v", err)
		}
		yamls = append(yamls, string(streamYaml))
	}

	return UnmarshalYaml(yamls...)
}

// DecodeBase64Strings will decode all the base64 strings supplied
func DecodeBase64Strings(values ...string) ([]string, error) {
	var contents []string
	for _, value := range values {
		content, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, err
		}
		contents = append(contents, string(content))
	}
	return contents, nil
}

// Fill will fill a according to DecoderConfig with the values from conf.
func Fill(keyPath string, conf interface{}, decoderConfig *mapstructure.DecoderConfig) error {
	value, err := GetValue(conf, keyPath)
	if err != nil {
		return err
	}

	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return fmt.Errorf("cant create decoder: %v", err)
	}

	err = decoder.Decode(value)
	if err != nil {
		return fmt.Errorf("failed to decode into `out`: %v", err)
	}

	return nil
}

// FillValue will fill a struct, out, with values from conf.
func FillValue(keyPath string, conf interface{}, out interface{}) bool {
	err := Fill(keyPath, conf, &mapstructure.DecoderConfig{Result: out})
	if err == nil {
		return true
	}
	return false
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

		switch reflect.ValueOf(value).Kind() {
		case reflect.Map:
			var ok bool
			value, ok = value.(map[interface{}]interface{})[part]
			if !ok {
				return nil, fmt.Errorf(
					"value at [%v] does not exist",
					currentPath)
			}
		case reflect.Slice:
			i, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf(
					"value at [%v] is array, but index [%v] is not int: %v",
					path.Dir(currentPath),
					part,
					err)
			}
			value = value.([]interface{})[i]
		default:
			return nil, fmt.Errorf(
				"value at [%v] not a map or slice: %v",
				part,
				reflect.ValueOf(value).Kind())
		}
	}
	return value, nil
}

// LoadConf will load all configurations provided.  In order of precedence
// (highest last), files, overrides.
func LoadConf(files []string, overrides []string) (map[interface{}]interface{}, error) {
	return ConfSources{Files: files, Overrides: overrides}.Load()

}

// LoadConfFromEnvironment will load all configurations present.  In order
// of precedence (highest last), files, YAML_FILES env var, overrides,
// YAML_VARS env var.
func LoadConfFromEnvironment(files []string, overrides []string) (map[interface{}]interface{}, error) {
	return ConfSources{Files: files, Overrides: overrides, Environment: true}.Load()
}

// LoadSettableConfFromEnvironment loads configuration for setting.  Only one
// file is allowed, but can be specified, either by the environment variable
// YAML_FILES, or as the single value in the supplied files array.  Returns
// the name of the file to be written, the conf map, and a non-nil error upon
// failure.  If the file does not currently exist, an empty map will be returned
// and a call to SaveConf will create the file.
func LoadSettableConfFromEnvironment(files []string) (string, map[interface{}]interface{}, error) {
	if yamlFiles, ok := os.LookupEnv("YAML_FILES"); ok {
		files = append(files, Splitter.Split(yamlFiles, -1)...)
	}
	if len(files) != 1 {
		return "", nil, errors.New("Exactly one file required with setv")
	}

	if _, err := os.Stat(files[0]); os.IsNotExist(err) {
		return files[0], map[interface{}]interface{}{}, nil
	}

	config, err := LoadConf(files, []string{})
	return files[0], config, err
}

// MarshalYaml will convert an object to yaml
func MarshalYaml(in interface{}) ([]byte, error) {
	value, err := yaml.Marshal(in)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// ReadEnvVars will read all the environment variables named and return an
// array of their values.  The order of the names to values will be
// preserved.
func ReadEnvVars(names ...string) ([]string, error) {
	var values []string
	for _, name := range names {
		if value, ok := os.LookupEnv(name); ok {
			values = append(values, value)
		} else {
			return nil, fmt.Errorf("Read env var [%s] failed, does not exist", name)
		}
	}
	return values, nil
}

// ReadFiles will read all the files supplied and return an array of their
// contents.  The order of files to contents will be preserved.
func ReadFiles(files ...string) ([]string, error) {
	var contents []string
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return nil, err
		}

		content, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		contents = append(contents, string(content))
	}
	return contents, nil
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

	return mergo.Merge(
		&parent,
		value,
		func(c *mergo.Config) { c.Overwrite = overwrite })
}

func getParentAndKey(config interface{}, keyPath string) (map[interface{}]interface{}, string, error) {
	configMap, ok := config.(map[interface{}]interface{})
	if !ok {
		return nil, "", fmt.Errorf("Config not a map")
	}
	parentParts, key := splitKeyPath(keyPath)
	if key == "" {
		return configMap, "", nil
	}

	parent := configMap
	for _, parentPart := range parentParts {
		parentValue, ok := parent[parentPart]
		if !ok {
			parentValue = make(map[interface{}]interface{})
			parent[parentPart] = parentValue
		}
		valueMap, ok := parentValue.(map[interface{}]interface{})
		if !ok {
			return nil, "", fmt.Errorf("Parent not a map")
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
	yamlBytes, err := MarshalYaml(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, yamlBytes, 0660)
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

	if err == io.EOF {
		return results, nil
	}
	return results, err
}

// UnmarshalYaml will parse all the supplied yaml strings, merge the resulting
// objects, and return the resulting map.  This only works for objects because
// the merge requires objects.
func UnmarshalYaml(yamlStrings ...string) (map[interface{}]interface{}, error) {
	var allYamls []interface{}
	for _, yamlString := range yamlStrings {
		yamls, err := UnmarshalAllYaml(yamlString)
		if err != nil {
			return nil, err
		}
		if len(yamls) > 0 {
			allYamls = append(allYamls, yamls...)
		}
	}

	result := make(map[interface{}]interface{})
	for index := len(allYamls) - 1; index >= 0; index-- {
		if err := mergo.Merge(&result, allYamls[index]); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// Walk will recursively iterate over all the nodes of conf calling callback
// for each node.
func Walk(callback func(key []string, value interface{}), conf interface{}) {
	node, ok := conf.(map[interface{}]interface{})
	if !ok {
		callback([]string{}, conf)
	}
	walk(callback, node, []string{})
}

func walk(callback func(key []string, value interface{}), node interface{}, keyStack []string) {
	switch node.(type) {
	case map[interface{}]interface{}:
		for k, v := range node.(map[interface{}]interface{}) {
			keyStack := append(keyStack, fmt.Sprintf("%v", k))
			walk(callback, v, keyStack)
		}
	case []interface{}:
		for i, j := range node.([]interface{}) {
			keyStack := append(keyStack, strconv.Itoa(i))
			walk(callback, j, keyStack)
		}
	default:
		callback(keyStack, node)
	}
}
