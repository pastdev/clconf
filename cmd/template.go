package cmd

import (
	"fmt"
	"os"

	"github.com/pastdev/clconf/v2/clconf"
	"github.com/spf13/cobra"
)

type templateContext struct {
	*rootContext
	templateOptions clconf.TemplateOptions
	inPlace         bool
	unixDirMode     string
	unixFileMode    string
}

var templateCmdContext = &templateContext{
	templateOptions: clconf.TemplateOptions{},
	rootContext:     rootCmdContext,
}

var templateCmd = &cobra.Command{
	Use:   "template <src1> [src2...] [destination folder]",
	Short: "Interpret a set of pre-existing templates",
	Long: `This will take an arbitrary number of source templates (or folders full
		of templates) and process them either in place (see --in-place) or into the
		folder specified as the last argument. It will make any folders required
		along the way. If a source is an existing file (not a folder) it will be
		treated as a template regardless of the extension (though if the extension
		matches it will still be removed).`,
	RunE: template,
	Example: `
	# Apply all templates with the .clconf extension to their relative folders in /dest
	template /tmp/srcFolder1 /tmp/srcFolder2 /dest

	# Apply all templates in both folders with the .clconf extension to the root of /dest
	template /tmp/srcFolder1 /tmp/srcFolder2 /dest --flatten

	# Interpret /tmp/srcFile.sh where it is (result is /tmp/srcFile.sh)
	template /tmp/srcFile.sh --in-place

	# Interpret /tmp/srcFile.sh.clconf where it is (result is /tmp/srcFile.sh)
	template /tmp/srcFile.sh.clconf --in-place

	# Interpret /tmp/srcFile.sh.clconf where it is (result is /tmp/srcFile.sh.clconf)
	template /tmp/srcFile.sh.clconf --in-place --template-extension ""
	`,
}

func init() {
	rootCmd.AddCommand(templateCmd)

	templateCmd.Flags().StringVar(&templateCmdContext.templateOptions.Extension, "template-extension", ".clconf",
		"Template file extension (will be removed during templating).")
	templateCmd.Flags().StringVar(&templateCmdContext.unixFileMode, "file-mode", "",
		"Chmod mode (e.g. 644) to apply to files when templating (new and existing) (defaults to copy from source template).")
	templateCmd.Flags().BoolVar(&templateCmdContext.templateOptions.KeepExistingPerms, "keep-existing-permissions", false,
		"Only apply --file-mode to new files, leave existing files as-is.")
	templateCmd.Flags().StringVar(&templateCmdContext.unixDirMode, "dir-mode", "775",
		"Chmod mode (e.g. 755) to apply to newly created directories.")
	templateCmd.Flags().BoolVar(&templateCmdContext.inPlace, "in-place", false,
		"Template the files in the folder they're found (implies no destination)")
	templateCmd.Flags().BoolVar(&templateCmdContext.templateOptions.Rm, "rm", false,
		"Remove template files after processing.")
	templateCmd.Flags().BoolVar(&templateCmdContext.templateOptions.Flatten, "flatten", false,
		"Don't preserve relative folders when processing a source folder.")
}

func template(cmd *cobra.Command, args []string) error {
	var dest string
	if !templateCmdContext.inPlace {
		if len(args) < 2 {
			return fmt.Errorf("Need at least two arguments when not using --in-place")
		}
		dest = args[len(args)-1]
		args = args[:len(args)-1]
	}

	mode, err := clconf.UnixModeToFileMode(templateCmdContext.unixDirMode)
	if err != nil {
		return fmt.Errorf("Error parsing dir-mode: %v", err)
	}

	templateCmdContext.templateOptions.DirMode = mode

	if templateCmdContext.unixFileMode == "" {
		templateCmdContext.templateOptions.CopyTemplatePerms = true
	} else {
		mode, err := clconf.UnixModeToFileMode(templateCmdContext.unixFileMode)
		if err != nil {
			return fmt.Errorf("Error parsing file-mode: %v", err)
		}

		templateCmdContext.templateOptions.CopyTemplatePerms = false
		templateCmdContext.templateOptions.FileMode = mode
	}

	if len(args) < 1 {
		return fmt.Errorf("No sources to process")
	}

	secretAgent, _ := templateCmdContext.newSecretAgent()
	value, err := templateCmdContext.getValue("/")
	if err != nil {
		return err
	}

	results, err := clconf.ProcessTemplates(args, dest, value, *secretAgent,
		templateCmdContext.templateOptions)
	if err != nil {
		return err
	}

	for _, result := range results {
		fmt.Fprintf(os.Stderr, "Templated: %q => %q\n", result.Src, result.Dest)
	}
	return nil
}
