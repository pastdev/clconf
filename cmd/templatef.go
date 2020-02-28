package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pastdev/clconf/v2/clconf"
	"github.com/spf13/cobra"
)

const setuid = 4
const setgid = 2
const sticky = 1

var templatefCmdContext = &templatefContext{
	rootContext: rootCmdContext,
}

type templatefContext struct {
	*rootContext
	inPlace           bool
	flatten           bool
	keepExistingPerms bool
	rm                bool
	fileMode          string
	dirMode           string
	extension         string
}

type pathWithRelative struct {
	fullPath string
	relPath  string
}

var templatefCmd = &cobra.Command{
	Use:   "templatef <src1> [src2...] [destination folder]",
	Short: "Interpret a set of pre-existing templates",
	Long: `This will take an arbitrary number of source templates (or folders full
		of templates and process them either in place (see --in-place) or into the
		folder specified as the last argument. It will make any folders required
		along the way. If a source is an existing file (not a folder) it will be
		treated as a template regardless of the extension (though if the extension
		matches it will still be removed).`,
	RunE: templatef,
	Example: `
	# Apply all templates with the .clconf extension to their relative folders in /dest
	templatef /tmp/srcFolder1 /tmp/srcFolder2 /dest

	# Apply all templates in both folders with the .clconf extension to the root of /dest
	templatef /tmp/srcFolder1 /tmp/srcFolder2 /dest --flatten

	# Interpret /tmp/srcFile.sh where it is (result is /tmp/srcFile.sh)
	templatef /tmp/srcFile.sh --in-place

	# Interpret /tmp/srcFile.sh.clconf where it is (result is /tmp/srcFile.sh)
	templatef /tmp/srcFile.sh.clconf --in-place

	# Interpret /tmp/srcFile.sh.clconf where it is (result is /tmp/srcFile.sh.clconf)
	templatef /tmp/srcFile.sh.clconf --in-place --template-extension ""
	`,
}

func init() {
	rootCmd.AddCommand(templatefCmd)

	templatefCmd.Flags().StringVar(&templatefCmdContext.extension, "template-extension", ".clconf",
		"Template file extension (will be removed during templating).")
	templatefCmd.Flags().StringVar(&templatefCmdContext.fileMode, "file-mode", "",
		"Chmod mode (e.g. 644) to apply to files when templating (new and existing) (defaults to copy from source template).")
	templatefCmd.Flags().BoolVar(&templatefCmdContext.keepExistingPerms, "keep-existing-permissions", false,
		"Only apply --file-mode to new files, leave existing files as-is.")
	templatefCmd.Flags().StringVar(&templatefCmdContext.dirMode, "dir-mode", "775",
		"Chmod mode (e.g. 755) to apply to newly created directories.")
	templatefCmd.Flags().BoolVar(&templatefCmdContext.inPlace, "in-place", false,
		"Template the files in the folder they're found (implies no destination)")
	templatefCmd.Flags().BoolVar(&templatefCmdContext.rm, "rm", false,
		"Remove template files after processing.")
	templatefCmd.Flags().BoolVar(&templatefCmdContext.rm, "flatten", false,
		"Don't preserve relative folders when processing a source folder.")
}

func templatef(cmd *cobra.Command, args []string) error {
	var dest string
	if !templatefCmdContext.inPlace {
		if len(args) < 2 {
			return fmt.Errorf("Need at least two arguments when not using --in-place")
		}
		dest = args[len(args)-1]
		args = args[:len(args)-1]
		perm, err := unixModeToFileMode(templatefCmdContext.dirMode)
		if err != nil {
			return err
		}

		err = os.MkdirAll(dest, perm)
		if err != nil {
			return err
		}
	}

	if len(args) < 1 {
		return fmt.Errorf("No sources to process")
	}

	secretAgent, _ := templatefCmdContext.newSecretAgent()
	value, err := templatefCmdContext.getValue("/")
	if err != nil {
		return err
	}

	return templatefCmdContext.processTemplates(args, dest, value, secretAgent)
}
func (c *templatefContext) processTemplates(srcs []string, dest string, value interface{}, secretAgent *clconf.SecretAgent) error {
	var templates []pathWithRelative
	for _, templateSrc := range srcs {
		srcTemplates, err := findTemplates(templateSrc, c.extension)
		if err != nil {
			return err
		}
		templates = append(templates, srcTemplates...)
	}

	for _, template := range templates {
		err := c.processTemplate(template, dest, value, secretAgent)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *templatefContext) processTemplate(paths pathWithRelative, dest string, value interface{}, secretAgent *clconf.SecretAgent) error {
	var mode os.FileMode
	var err error

	if c.fileMode != "" {
		mode, err = unixModeToFileMode(c.fileMode)
		if err != nil {
			return err
		}
	} else {
		stat, err := os.Stat(paths.fullPath)
		if err != nil {
			return err
		}
		mode = stat.Mode()
	}

	var target = paths.fullPath
	if dest != "" {
		if c.flatten {
			target = path.Join(dest, path.Base(target))
		} else {
			target = path.Join(dest, paths.relPath)
		}
	}

	target = strings.TrimSuffix(target, c.extension)

	targetDir := filepath.Dir(target)
	dirPerm, err := unixModeToFileMode(c.dirMode)
	if err != nil {
		return err
	}
	err = os.MkdirAll(targetDir, dirPerm)
	if err != nil {
		return fmt.Errorf("Error making target dir %q: %v", targetDir, err)
	}

	template, err := clconf.NewTemplateFromFile("cli", paths.fullPath,
		&clconf.TemplateConfig{
			SecretAgent: secretAgent,
		})
	if err != nil {
		return err
	}

	content, err := template.Execute(value)
	if err != nil {
		return fmt.Errorf("Error processing template: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Templating: %q => %q", paths.fullPath, target)
	err = ioutil.WriteFile(target, []byte(content), mode)
	if err != nil {
		return err
	}

	if !c.keepExistingPerms {
		os.Chmod(target, mode)
	}

	if c.rm && paths.fullPath != target {
		err = os.Remove(paths.fullPath)
		if err != nil {
			return err
		}
	}
	return nil
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
			result = append(result, pathWithRelative{relPath: relPath, fullPath: path})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func unixModeToFileMode(unixMode string) (os.FileMode, error) {
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
