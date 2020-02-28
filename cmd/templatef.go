package cmd

import (
	"fmt"

	"github.com/pastdev/clconf/v2/clconf"
	"github.com/spf13/cobra"
)

type templatefContext struct {
	*rootContext
	templateSettings clconf.TemplateSettings
	inPlace          bool
}

var templatefCmdContext = &templatefContext{
	templateSettings: clconf.TemplateSettings{},
	rootContext:      rootCmdContext,
}

var templatefCmd = &cobra.Command{
	Use:   "templatef <src1> [src2...] [destination folder]",
	Short: "Interpret a set of pre-existing templates",
	Long: `This will take an arbitrary number of source templates (or folders full
		of templates) and process them either in place (see --in-place) or into the
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

	templatefCmd.Flags().StringVar(&templatefCmdContext.templateSettings.Extension, "template-extension", ".clconf",
		"Template file extension (will be removed during templating).")
	templatefCmd.Flags().StringVar(&templatefCmdContext.templateSettings.FileMode, "file-mode", "",
		"Chmod mode (e.g. 644) to apply to files when templating (new and existing) (defaults to copy from source template).")
	templatefCmd.Flags().BoolVar(&templatefCmdContext.templateSettings.KeepExistingPerms, "keep-existing-permissions", false,
		"Only apply --file-mode to new files, leave existing files as-is.")
	templatefCmd.Flags().StringVar(&templatefCmdContext.templateSettings.DirMode, "dir-mode", "775",
		"Chmod mode (e.g. 755) to apply to newly created directories.")
	templatefCmd.Flags().BoolVar(&templatefCmdContext.inPlace, "in-place", false,
		"Template the files in the folder they're found (implies no destination)")
	templatefCmd.Flags().BoolVar(&templatefCmdContext.templateSettings.Rm, "rm", false,
		"Remove template files after processing.")
	templatefCmd.Flags().BoolVar(&templatefCmdContext.templateSettings.Flatten, "flatten", false,
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
	}

	if len(args) < 1 {
		return fmt.Errorf("No sources to process")
	}

	secretAgent, _ := templatefCmdContext.newSecretAgent()
	value, err := templatefCmdContext.getValue("/")
	if err != nil {
		return err
	}

	return templatefCmdContext.templateSettings.ProcessTemplates(args, dest, value, secretAgent)
}
