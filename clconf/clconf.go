// Package clconf provides functions to extract values from a set of yaml
// files after merging them.
package clconf

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/imdario/mergo"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var splitter = regexp.MustCompile(`,`)

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

// FillValue will fill a struct, out, with values from conf.
func FillValue(keyPath string, conf interface{}, out interface{}) bool {
	value, ok := GetValue(keyPath, conf)
	if !ok {
		return false
	}
	err := mapstructure.Decode(value, out)
	if err != nil {
		return false
	}
	return ok
}

// GetValue returns the value at the indicated path.  Paths are separated by
// the '/' character.
func GetValue(keyPath string, conf interface{}) (interface{}, bool) {
	if keyPath == "" {
		return conf, true
	}

	var value = conf
	for _, part := range strings.Split(keyPath, "/") {
		if part == "" {
			continue
		}
		if reflect.ValueOf(value).Kind() != reflect.Map {
			log.Warnf("value at [%v] not a map: %v", part, reflect.ValueOf(value).Kind())
			return nil, false
		}
		partValue, ok := value.(map[interface{}]interface{})[part]
		if !ok {
			log.Warnf("value at [%v] does not exist", part)
			return nil, false
		}
		value = partValue
	}
	return value, true
}

// LoadConf will load all configurations provided.  In order of precedence
// (highest last), files, overrides.
func LoadConf(files []string, overrides []string) (map[interface{}]interface{}, error) {
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

	return UnmarshalYaml(yamls...)
}

// LoadConfFromEnvironment will load all configurations present.  In order
// of precedence (highest last), YAML_FILES env var, YAML_VARS env var,
// files, overrides.
func LoadConfFromEnvironment(files []string, overrides []string) (map[interface{}]interface{}, error) {
	if yamlFiles, ok := os.LookupEnv("YAML_FILES"); ok {
		files = append(files, splitter.Split(yamlFiles, -1)...)
	}
	if yamlVars, ok := os.LookupEnv("YAML_VARS"); ok {
		overrides = append(overrides, ReadEnvVars(splitter.Split(yamlVars, -1)...)...)
	}
	return LoadConf(files, overrides)
}

// LoadSettableConfFromEnvironment loads configuration for setting.  Only one
// file is allowed, but can be specified, either by the environment variable
// YAML_FILES, or as the single value in the supplied files array.  Returns
// the name of the file to be written, the conf map, and a non-nil error upon
// failure.
func LoadSettableConfFromEnvironment(files []string) (string, map[interface{}]interface{}, error) {
	if yamlFiles, ok := os.LookupEnv("YAML_FILES"); ok {
		files = append(files, splitter.Split(yamlFiles, -1)...)
	}
	if len(files) > 1 {
		return "", nil, errors.New("Only one file allowed with setv")
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
func ReadEnvVars(names ...string) []string {
	var values []string
	for _, name := range names {
		if value, ok := os.LookupEnv(name); ok {
			values = append(values, value)
		} else {
			log.Panicf("Read env var [%s] failed, does not exist", name)
		}
	}
	return values
}

// ReadFiles will read all the files supplied and return an array of their
// contents.  The order of files to contents will be preserved.
func ReadFiles(files ...string) ([]string, error) {
	var contents []string
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return nil, err
			//log.Panicf("Read file [%s] failed, does not exist", file)
		}

		content, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
			//log.Panicf("Read file [%s] failed: %v", file, err)
		}
		contents = append(contents, string(content))
	}
	return contents, nil
}

func splitKeyPath(keyPath string) ([]string, string) {
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

func SetValue(config map[interface{}]interface{}, keyPath string, value interface{}) error {
	parentParts, key := splitKeyPath(keyPath)
	if key == "" {
	    return fmt.Errorf("[%v] is an invalid keyPath", keyPath)
	}

	parent := config
	for _, parentPart := range parentParts {
		parentValue, ok := parent[parentPart]
		if !ok {
			parentValue = make(map[interface{}]interface{})
			parent[parentPart] = parentValue
		}
		valueMap, ok := parentValue.(map[interface{}]interface{})
		if !ok {
			return fmt.Errorf("Parent not a map")
		}

		parent = valueMap
	}

	parent[key] = value

	return nil
}

func unmarshalYaml(yamlBytes ...[]byte) (map[interface{}]interface{}, error) {
	result := make(map[interface{}]interface{})
	for index := len(yamlBytes) - 1; index >= 0; index-- {
		yamlMap := make(map[interface{}]interface{})

		err := yaml.Unmarshal(yamlBytes[index], &yamlMap)
		if err != nil {
			return nil, err
		}

		if err := mergo.Merge(&result, yamlMap); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func SaveConf(file string, config map[interface{}]interface{}) error {
	yamlBytes, err := MarshalYaml(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, yamlBytes, 0660)
}

// UnmarshalYaml will parse all the supplied yaml strings, merge the resulting
// objects, and return the resulting map
func UnmarshalYaml(yamlStrings ...string) (map[interface{}]interface{}, error) {
	yamlBytes := make([][]byte, len(yamlStrings))
	for _, yaml := range yamlStrings {
		yamlBytes = append(yamlBytes, []byte(yaml))
	}
	return unmarshalYaml(yamlBytes...)
}
