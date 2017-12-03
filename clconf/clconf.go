// Package clconf provides functions to extract values from a set of yaml
// files after merging them.
package clconf

import (
	"encoding/base64"
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

var splitter = regexp.MustCompile(`\s+`)

// DecodeBase64Strings will decode all the base64 strings supplied
func DecodeBase64Strings(values ...string) []string {
	var contents []string
	for _, value := range values {
		content, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			log.Panicf("Base64 parsing failed: %v", err)
		}
		contents = append(contents, string(content))
	}
	return contents
}

// FillValue will fill a struct, out, with values from conf.
func FillValue(path string, conf interface{}, out interface{}) bool {
	value, ok := GetValue(path, conf)
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
func GetValue(path string, conf interface{}) (interface{}, bool) {
	if path == "" {
		return conf, true
	}

	var value = conf
	for _, part := range strings.Split(path, "/") {
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

// LoadConf will load all configurations present.  In order of precedence
// (highest last), YAML_FILES env var, YAML_VARS env var, overrides.
func LoadConf(overrides ...string) map[interface{}]interface{} {
	yamls := []string{}
	if yamlFiles, ok := os.LookupEnv("YAML_FILES"); ok {
		yamls = append(yamls,
			ReadFiles(
				splitter.Split(yamlFiles, -1)...)...)
	}
	if yamlVars, ok := os.LookupEnv("YAML_VARS"); ok {
		yamls = append(yamls,
			DecodeBase64Strings(
				ReadEnvVars(
					splitter.Split(yamlVars, -1)...)...)...)
	}
	if len(overrides) > 0 {
		yamls = append(yamls, DecodeBase64Strings(overrides...)...)
	}

	conf, err := UnmarshalYaml(yamls...)
	if err != nil {
		log.Panicf("Load conf failed: %v", err)
	}
	return conf
}

// MarshalYaml will convert an object to yaml
func MarshalYaml(in interface{}) (string, error) {
	value, err := yaml.Marshal(in)
	if err != nil {
		return "", err
	}
	return string(value), nil
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
func ReadFiles(files ...string) []string {
	var contents []string
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			log.Panicf("Read file [%s] failed, does not exist", file)
		}

		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.Panicf("Read file [%s] failed: %v", file, err)
		}
		contents = append(contents, string(content))
	}
	return contents
}

// UnmarshalYaml will parse all the supplied yaml strings, merge the resulting
// objects, and return the resulting map
func UnmarshalYaml(yamlStrings ...string) (map[interface{}]interface{}, error) {
	result := make(map[interface{}]interface{})
	for index := len(yamlStrings) - 1; index >= 0; index-- {
		yamlMap := make(map[interface{}]interface{})

		err := yaml.Unmarshal([]byte(yamlStrings[index]), &yamlMap)
		if err != nil {
			log.Warnf("error in yaml [%d]: %v", index, err)
			return nil, err
		}

		if err := mergo.Merge(&result, yamlMap); err != nil {
			log.Panicf("error at index [%d] yaml: %v", index, err)
			return nil, err
		}
	}
	return result, nil
}
