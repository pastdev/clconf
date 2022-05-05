package yamljson

import (
	"encoding/json"
	"fmt"
	"os"

	jsonpatch "github.com/evanphx/json-patch/v5"
)

// PatchFromFiles applies a list of rfc 6902 patches to the data.
func PatchFromFiles(data interface{}, patchFiles ...string) (interface{}, error) {
	patches := make([][]byte, len(patchFiles))
	var err error
	for i, patch := range patchFiles {
		patches[i], err = os.ReadFile(patch)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", patch, err)
		}
	}
	return Patch(data, patches...)
}

// PatchFromStrings applies a list of rfc 6902 patches to the data.
func PatchFromStrings(data interface{}, patchStrings ...string) (interface{}, error) {
	patches := make([][]byte, len(patchStrings))
	for i, patch := range patchStrings {
		patches[i] = []byte(patch)
	}
	return Patch(data, patches...)
}

// Patch applies a list of rfc 6902 patches to the data.
func Patch(data interface{}, patchBytes ...[]byte) (interface{}, error) {
	patches := make([]jsonpatch.Patch, len(patchBytes))
	for i, patch := range patchBytes {
		// same approach as jsonpatch.DecodePatch, but need to do it
		// ourselves in case the source is yaml instead of json.
		var decoded jsonpatch.Patch
		patch, err := YAMLToJSON(patch)
		if err != nil {
			return nil, fmt.Errorf("yaml to json %s: %w", patch, err)
		}
		err = json.Unmarshal(patch, &decoded)
		if err != nil {
			return nil, fmt.Errorf("decoding %s: %w", patch, err)
		}
		patches[i] = decoded
	}

	patched, err := json.Marshal(ConvertMapIToMapS(data))
	if err != nil {
		return nil, fmt.Errorf("converting yaml to json: %w", err)
	}

	for _, p := range patches {
		patched, err = p.Apply(patched)
		if err != nil {
			return nil, fmt.Errorf("converting yaml to json: %w", err)
		}
	}

	data, err = UnmarshalYamlInterface(string(patched))
	if err != nil {
		return nil, fmt.Errorf("umarshal patched: %w", err)
	}

	return ConvertMapSToMapI(data), nil
}
