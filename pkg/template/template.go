package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pastdev/clconf/v3/pkg/secret"
)

const setuid = 4
const setgid = 2
const sticky = 1

// TemplateOptions are settings for ProcessTemplates.
type TemplateOptions struct { //nolint:revive
	// CopyTemplatePerms uses the existing template permissions instead of FileMode for template
	// permissions
	CopyTemplatePerms bool
	// Flatten flattens the templates into the root of the dest instead of the preserving
	// the relative path under the source.
	Flatten bool
	// KeepEmpty forces empty result files to be written (usually not written or removed if already existing)
	KeepEmpty bool
	// KeepExistingPerms determines whether to use existing permissions on template files that are overwritten.
	KeepExistingPerms bool
	// Rm determines whether template files distinct from their target are deleted after processing.
	Rm bool
	// FileMode is the permissions to apply to template files when writing.
	FileMode os.FileMode
	// DirMode is the permission similar to FileMode but for new folders.
	DirMode os.FileMode
	// Extension is the extension to use when searching folders. If missing all files will be used.
	// The extension is stripped from the file name when templating.
	Extension string
	// LeftDelim is passed to go teplate.Delims
	LeftDelim string
	// RightDelim is passed to go teplate.Delims
	RightDelim string
}

type pathWithRelative struct {
	full string
	rel  string
}

// TemplateResult stores the result of a single template processing.
type TemplateResult struct { //nolint:revive
	Src  string
	Dest string
}

// ProcessTemplates processes templates. If dest is non empty it must be a folder into which
// templates will be placed after processing (the folder will be created if necessary). If empty
// templates are processed into the folders in which they are found.
func ProcessTemplates(srcs []string, dest string, value interface{}, secretAgent *secret.SecretAgent,
	options TemplateOptions,
) ([]TemplateResult, error) {
	if dest != "" {
		err := MkdirAllNoUmask(dest, options.DirMode)
		if err != nil {
			return nil, err
		}
	}

	results := []TemplateResult{}

	for _, templateSrc := range srcs {
		templates, err := findTemplates(templateSrc, options.Extension)
		if err != nil {
			return nil, err
		}

		for _, template := range templates {
			result, err := processTemplate(template, dest, value, secretAgent, options)
			if err != nil {
				return nil, err
			}
			results = append(results, result)
		}
	}
	return results, nil
}

func processTemplate(
	paths pathWithRelative,
	dest string,
	value interface{},
	secretAgent *secret.SecretAgent,
	options TemplateOptions,
) (TemplateResult, error) {
	var mode os.FileMode
	var err error

	result := TemplateResult{}

	if options.CopyTemplatePerms {
		stat, err := os.Stat(paths.full)
		if err != nil {
			return result, fmt.Errorf("stat for copy perms: %w", err)
		}
		mode = stat.Mode()
	} else {
		mode = options.FileMode
	}

	var target = paths.full
	if dest != "" {
		if options.Flatten {
			target = filepath.Join(dest, filepath.Base(target))
		} else {
			target = filepath.Join(dest, paths.rel)
		}
	}

	target = strings.TrimSuffix(target, options.Extension)

	targetDir := filepath.Dir(target)
	err = MkdirAllNoUmask(targetDir, options.DirMode)
	if err != nil {
		return result, fmt.Errorf("making target dir %q: %w", targetDir, err)
	}

	template, err := NewTemplateFromFile(paths.rel, paths.full,
		&TemplateConfig{
			SecretAgent: secretAgent,
			LeftDelim:   options.LeftDelim,
			RightDelim:  options.RightDelim,
		})
	if err != nil {
		return result, err
	}

	content, err := template.Execute(value)
	if err != nil {
		return result, fmt.Errorf("processing template: %w", err)
	}

	if options.KeepEmpty || content != "" {
		err = os.WriteFile(target, []byte(content), mode)
		if err != nil {
			return result, fmt.Errorf("write target: %w", err)
		}

		if !options.KeepExistingPerms {
			// WriteFile only uses the mode if the file is created during write
			// and even then filters on the umask so we os.Chmod to get the
			// requested perms.
			err = os.Chmod(target, mode)
			if err != nil {
				return result, fmt.Errorf("chmod target: %w", err)
			}
		}
	} else { // Don't keep the empty result
		_, err = os.Stat(target)
		if err == nil {
			err = os.Remove(target)
		}
		if err != nil && !os.IsNotExist(err) {
			return result, fmt.Errorf("remove empty: %w", err)
		}
	}

	if options.Rm && paths.full != target {
		err = os.Remove(paths.full)
		if err != nil {
			return result, fmt.Errorf("remove template: %w", err)
		}
	}

	result.Dest = target
	result.Src = paths.full

	return result, nil
}

// findTemplates returns the templates under the given path as strings in the format
// <relativePath> + os.PathListSeparator + <fullPath>
func findTemplates(startPath string, extension string) ([]pathWithRelative, error) {
	var result []pathWithRelative
	startPath = filepath.Clean(startPath)

	err := filepath.Walk(startPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, extension) || path == startPath) {
			relPath, err := filepath.Rel(startPath, path)
			if err != nil {
				return fmt.Errorf("reltative path: %w", err)
			}
			if strings.HasPrefix(relPath, "..") {
				return fmt.Errorf("path %q somehow ended up outside starting path %q (relpath: %q)",
					path, startPath, relPath)
			}
			if relPath == "." {
				relPath = filepath.Base(path)
			}
			result = append(result, pathWithRelative{rel: relPath, full: path})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("find templates: %w", err)
	}
	return result, nil
}

// UnixModeToFileMode converts a unix file mode including special bits to a golang os.FileMode.
// The special bits (sticky, setuid, setgid) don't line up exactly between the two.
// Example: 02777 would set the setuid bit on unix but would end up 0777 if used as an os.FileMode
func UnixModeToFileMode(unixMode string) (os.FileMode, error) {
	intVal, err := strconv.ParseInt(unixMode, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("parse file mode int: %w", err)
	}

	fileMode := os.FileMode(uint32(intVal)) //nolint: gosec // this value is parsed right above so will not overflow
	perms := uint8((intVal) >> 9)           //nolint: gosec // this value is parsed right above so will not overflow

	if perms&sticky != 0 {
		fileMode |= os.ModeSticky
	}
	if perms&setgid != 0 {
		fileMode |= os.ModeSetgid
	}
	if perms&setuid != 0 {
		fileMode |= os.ModeSetuid
	}

	return fileMode, nil
}
