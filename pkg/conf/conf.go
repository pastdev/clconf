package conf

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/pastdev/clconf/v3/pkg/yamljson"
)

// Splitter is the regex used to split YAML_FILES and YAML_VARS
var Splitter = regexp.MustCompile(`,`)

// ConfSources contains sources of yaml for loading. See Load() for precedence
type ConfSources struct { //nolint:revive
	// Environment loads config from environment vars when true. The vars loaded
	// are:
	// YAML_FILES: comma separated values will be appended to Files
	// YAML_VARS: comma separated values of other environment variables to read
	// and whose base64 strings will be appended to Overrides
	Environment bool
	// Files is a list of filenames to read
	Files []string
	// Overrides are Base64 encoded strings of yaml
	Overrides []string
	// Patches are files containing JSON 6902 patches to apply after the merge
	// is complete
	Patches []string
	// PatchStrings are strings containing JSON 6902 patches to apply after the
	// merge is complete
	PatchStrings []string
	// An optional (can be nil) stream to read raw yaml (potentially multiple
	// inline documents)
	Stream io.Reader
}

// LoadInterface will load the config determined by settings in the struct. In order
// of precedence (highest last), Files, YAML_FILES env var, Overrides,
// YAML_VARS env var, Stream.
func (s ConfSources) LoadInterface() (interface{}, error) {
	conf, _, err := s.loadInterface(false)
	return conf, err
}

// LoadSettableInterface will load the config determined by settings in the
// struct. Will fail if more than one file source or any non-file source is
// indicated. Returns the conf, and the file that values can be set in.
func (s ConfSources) LoadSettableInterface() (interface{}, string, error) {
	return s.loadInterface(true)
}

// LoadInterface will load the config determined by settings in the struct. In order
// of precedence (highest last), Files, YAML_FILES env var, Overrides,
// YAML_VARS env var, Stream.
func (s ConfSources) loadInterface(settable bool) (interface{}, string, error) {
	files := s.Files
	overrides := s.Overrides

	if s.Environment {
		if yamlFiles, ok := os.LookupEnv("YAML_FILES"); ok && len(yamlFiles) > 0 {
			files = append(files, Splitter.Split(yamlFiles, -1)...)
		}
		if yamlVars, ok := os.LookupEnv("YAML_VARS"); ok && len(yamlVars) > 0 {
			envVars, err := ReadEnvVars(Splitter.Split(yamlVars, -1)...)
			if err != nil {
				return nil, "", err
			}
			overrides = append(overrides, envVars...)
		}
	}

	yamls := []string{}
	if len(files) > 0 {
		if settable && len(files) > 1 {
			return nil, "", fmt.Errorf("only single file allowed when settable, found: %v", files)
		}
		moreYamls, err := ReadFiles(files...)
		if err != nil {
			return nil, "", err
		}
		yamls = append(yamls, moreYamls...)
	} else if settable {
		return nil, "", errors.New("settable requires single file")
	}

	if len(overrides) > 0 {
		if settable {
			return nil, "", errors.New("overrides not allowed when settable")
		}
		moreYamls, err := DecodeBase64Strings(overrides...)
		if err != nil {
			return nil, "", err
		}
		yamls = append(yamls, moreYamls...)
	}

	if s.Stream != nil {
		if settable {
			return nil, "", errors.New("stream not allowed when settable")
		}
		streamYaml, err := ioutil.ReadAll(s.Stream)
		if err != nil {
			return nil, "", fmt.Errorf("reading stdin: %w", err)
		}
		yamls = append(yamls, string(streamYaml))
	}

	merged, err := yamljson.UnmarshalYamlInterface(yamls...)
	if err != nil {
		return nil, "", fmt.Errorf("unmarshal: %w", err)
	}

	if len(s.Patches) > 0 {
		if settable {
			return nil, "", errors.New("patch not allowed when settable")
		}
		merged, err = yamljson.PatchFromFiles(merged, s.Patches...)
		if err != nil {
			return nil, "", fmt.Errorf("patch: %w", err)
		}
	}

	if len(s.PatchStrings) > 0 {
		if settable {
			return nil, "", errors.New("patch string not allowed when settable")
		}
		merged, err = yamljson.PatchFromStrings(merged, s.PatchStrings...)
		if err != nil {
			return nil, "", fmt.Errorf("patch string: %w", err)
		}
	}

	if settable {
		return merged, files[0], nil
	}

	return merged, "", nil
}

// DecodeBase64Strings will decode all the base64 strings supplied
func DecodeBase64Strings(values ...string) ([]string, error) {
	var contents []string
	for _, value := range values {
		content, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, fmt.Errorf("decode base64: %w", err)
		}
		contents = append(contents, string(content))
	}
	return contents, nil
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
			return nil, fmt.Errorf("read env var [%s] failed, does not exist", name)
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
			return nil, fmt.Errorf("stat: %w", err)
		}

		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read: %w", err)
		}
		contents = append(contents, string(content))
	}
	return contents, nil
}
