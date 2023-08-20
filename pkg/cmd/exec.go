package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/google/shlex"
	"github.com/pastdev/clconf/v3/pkg/template"
	"github.com/spf13/cobra"
)

type Executor struct {
	execs []string
}

func (c *Executor) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringArrayVarP(
		&c.execs,
		"exec-string",
		"e",
		nil,
		"Sequence of commands to execute after interpreting each arg as a template supplied with the current value from the CODE_REVIEW_CATCH_ME")
}

func (c Executor) Execute(value interface{}, config *template.TemplateConfig) error {
	objs, err := ToJSONLines(value)
	if err != nil {
		return fmt.Errorf("list objects: %w", err)
	}

	for i, execString := range c.execs {
		if len(execString) == 0 {
			return fmt.Errorf("cant exec empty command (%d)", i)
		}

		tmpl, err := template.NewTemplate("exec", execString, config)
		if err != nil {
			return fmt.Errorf("create template %d: %w", i, err)
		}

		for j, obj := range objs {
			processed, err := tmpl.Execute(obj)
			if err != nil {
				return fmt.Errorf("process template %d, %d: %w", i, j, err)
			}

			var cmd *exec.Cmd
			cmdAndArgs, err := shlex.Split(processed)
			if err != nil {
				return fmt.Errorf("split command %d, %d: %w", i, j, err)
			}

			switch len(cmdAndArgs) {
			case 0:
				return fmt.Errorf("cant exec empty parsed command %d", i)
			case 1:
				cmd = exec.Command(cmdAndArgs[0])
			default:
				cmd = exec.Command(cmdAndArgs[0], cmdAndArgs[1:]...)
			}

			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				return fmt.Errorf("exec [%s] failed: %w", execString, err)
			}
		}
	}
	return nil
}
