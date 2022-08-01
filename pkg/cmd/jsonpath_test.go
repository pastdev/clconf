package cmd

import (
	"testing"

	"github.com/pastdev/clconf/v3/pkg/yamljson"
	"github.com/stretchr/testify/require"
)

const jsonPathBuildYml = `---
version: 3
images:
- build_type: docker
  tag: example.org/image:latest
  dockerfile: docker/run/Dockerfile
- build_type: docker
  tag: example.org/otherimage:latest
  dockerfile: docker/test/Dockerfile
  publish:
  - type: image
`

func TestJSONPathParse(t *testing.T) {
	test := func(name, path string, first bool, expected string) {
		t.Run(name, func(t *testing.T) {
			data, err := yamljson.UnmarshalSingleYaml(jsonPathBuildYml)
			require.NoError(t, err, "Unmarshal succeeds")
			result, err := evaluateJSONPath(path, data, first)
			require.NoError(t, err, "evaluateJsonPath succeeds")
			actual, err := yamljson.MarshalYaml(result)
			require.NoError(t, err, "back to yaml succeeds")
			require.Equal(t, expected, string(actual), "results match expectations")
		})
	}

	test("get values", "$.images..tag", false,
		"- example.org/image:latest\n- example.org/otherimage:latest\n")
	test("get first value", "$.images..tag", true,
		"example.org/image:latest\n")
	test("get first complex as list", "$.images[0]", false,
		`- build_type: docker
  dockerfile: docker/run/Dockerfile
  tag: example.org/image:latest
`)
	test("get first complex as single", "$.images[0]", true,
		`build_type: docker
dockerfile: docker/run/Dockerfile
tag: example.org/image:latest
`)
	test("filter selection existance of child", "$.images[?(@.publish)].tag", false,
		"- example.org/otherimage:latest\n")
	test("filter selection on child content", "$.images[?(@.dockerfile == 'docker/run/Dockerfile')].tag", false,
		"- example.org/image:latest\n")
	test("filter selection on child content, first only", "$.images[?(@.dockerfile == 'docker/run/Dockerfile')].tag", true,
		"example.org/image:latest\n")
	test("filter selection on child content 2", "$.images[?(@.dockerfile == 'docker/test/Dockerfile')].tag", false,
		"- example.org/otherimage:latest\n")
	test("filter selection on child content 2, first only", "$.images[?(@.dockerfile == 'docker/test/Dockerfile')].tag", true,
		"example.org/otherimage:latest\n")
}
