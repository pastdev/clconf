package clconf

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const setuid = 4
const setgid = 2
const sticky = 1

// TemplateSettings are settings for ProcessTemplates.
type TemplateSettings struct {
	// Flatten determines if directory structure is preserved when scanning folders
	// for templates.
	Flatten bool
	// KeepExistingPerms determines whether to use existing permissions on template files that are overwritten.
	KeepExistingPerms bool
	// Rm determines whether template files distinct from their target are deleted after processing.
	Rm bool
	// FileMode is a string of unix file permissions (e.g. 2755) that apply to templates. An empty string
	// will copy permissions from the template itself.
	FileMode string
	// DirMode is the permission similar to FileMode but for new folders.
	DirMode string
	// Extension is the extension to use when searching folders. If missing all files will be used.
	// The extension is stripped from the file name when templating.
	Extension string
}

type pathWithRelative struct {
	full string
	rel  string
}

// TemplateResult stores the result of a single template processing.
type TemplateResult struct {
	Src  string
	Dest string
}

// ProcessTemplates processes templates. If dest is populated src files/folders are searched for
// templates and placed into dest. Otherwise files are replaced in the folders they are found.
func (c *TemplateSettings) ProcessTemplates(srcs []string, dest string, value interface{}, secretAgent *SecretAgent) ([]TemplateResult, error) {
	if dest != "" {
		perm, err := UnixModeToFileMode(c.DirMode)
		if err != nil {
			return nil, err
		}

		err = MkdirAllNoUmask(dest, perm)
		if err != nil {
			return nil, err
		}
	}

	results := []TemplateResult{}

	for _, templateSrc := range srcs {
		templates, err := findTemplates(templateSrc, c.Extension)
		if err != nil {
			return nil, err
		}

		for _, template := range templates {
			result, err := c.processTemplate(template, dest, value, secretAgent)
			if err != nil {
				return nil, err
			}
			results = append(results, result)
		}
	}
	return results, nil
}

func (c *TemplateSettings) processTemplate(paths pathWithRelative, dest string, value interface{}, secretAgent *SecretAgent) (TemplateResult, error) {
	var mode os.FileMode
	var err error

	result := TemplateResult{}

	if c.FileMode != "" {
		mode, err = UnixModeToFileMode(c.FileMode)
		if err != nil {
			return result, err
		}
	} else {
		stat, err := os.Stat(paths.full)
		if err != nil {
			return result, err
		}
		mode = stat.Mode()
	}

	var target = paths.full
	if dest != "" {
		if c.Flatten {
			target = path.Join(dest, path.Base(target))
		} else {
			target = path.Join(dest, paths.rel)
		}
	}

	target = strings.TrimSuffix(target, c.Extension)

	targetDir := filepath.Dir(target)
	dirPerm, err := UnixModeToFileMode(c.DirMode)
	if err != nil {
		return result, err
	}
	err = MkdirAllNoUmask(targetDir, dirPerm)
	if err != nil {
		return result, fmt.Errorf("Error making target dir %q: %v", targetDir, err)
	}

	template, err := NewTemplateFromFile(paths.rel, paths.full,
		&TemplateConfig{
			SecretAgent: secretAgent,
		})
	if err != nil {
		return result, err
	}

	content, err := template.Execute(value)
	if err != nil {
		return result, fmt.Errorf("Error processing template: %v", err)
	}

	err = ioutil.WriteFile(target, []byte(content), mode)
	if err != nil {
		return result, err
	}

	if !c.KeepExistingPerms {
		err = os.Chmod(target, mode)
		if err != nil {
			return result, err
		}
	}

	if c.Rm && paths.full != target {
		err = os.Remove(paths.full)
		if err != nil {
			return result, err
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
				return err
			}
			if strings.HasPrefix(relPath, "..") {
				return fmt.Errorf("path %q somehow ended up outside starting path %q (relpath: %q)", path, startPath, relPath)
			}
			if relPath == "." {
				relPath = filepath.Base(path)
			}
			result = append(result, pathWithRelative{rel: relPath, full: path})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, err
}

// MkdirAllNoUmask is os.MkdirAll that ignores the current unix umask.
func MkdirAllNoUmask(path string, perms os.FileMode) error {
	existing := syscall.Umask(0)
	defer syscall.Umask(existing)
	return os.MkdirAll(path, perms)
}

// UnixModeToFileMode converts a unix file mode including special bits to a golang os.FileMode.
// The bits don't line up natively.
func UnixModeToFileMode(unixMode string) (os.FileMode, error) {
	intVal, err := strconv.ParseInt(unixMode, 8, 32)
	if err != nil {
		return 0, err
	}

	fileMode := os.FileMode(uint32(intVal))
	perms := uint8((intVal) >> 9)
	if err != nil {
		return 0, err
	}

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
