package yamljson_test

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/pastdev/clconf/v3/pkg/yamljson"
	"github.com/stretchr/testify/assert"
)

func TestPatch(t *testing.T) {
	tester := func(
		name string,
		expected map[interface{}]interface{},
		data map[interface{}]interface{},
		patches ...string,
	) {
		t.Run(name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "clconf")
			if err != nil {
				t.Fatalf("create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tempDir) }()

			patchFiles := make([]string, len(patches))
			for i, patch := range patches {
				patchFiles[i] = path.Join(tempDir, fmt.Sprintf("file_%d.log", i))
				err := os.WriteFile(patchFiles[i], []byte(patch), 0600)
				if err != nil {
					t.Fatalf("write file: %v", err)
				}
			}

			actual, err := yamljson.PatchFromFiles(data, patchFiles...)
			assert.Nil(t, err)
			assert.Equal(t, expected, actual)

			actual, err = yamljson.PatchFromStrings(data, patches...)
			assert.Nil(t, err)
			assert.Equal(t, expected, actual)
		})
	}

	tester(
		"simple replace",
		map[interface{}]interface{}{"foo": "baz"},
		map[interface{}]interface{}{"foo": "bar"},
		`[{"op":"replace","path":"/foo","value":"baz"}]`,
	)
	tester(
		"multiple replace",
		map[interface{}]interface{}{"foo": "baz", "hip": "hap"},
		map[interface{}]interface{}{"foo": "bar", "hip": "hop"},
		`[{"op":"replace","path":"/foo","value":"baz"},{"op":"replace","path":"/hip","value":"hap"}]`,
	)
	tester(
		"yaml patch",
		map[interface{}]interface{}{"foo": "baz", "hip": "hap"},
		map[interface{}]interface{}{"foo": "bar", "hip": "hop"},
		`---
        - op: replace
          path: /foo
          value: baz
        - op: replace
          path: /hip
          value: hap
        `,
	)
}
