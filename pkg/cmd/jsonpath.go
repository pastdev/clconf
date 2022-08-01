package cmd

import (
	"fmt"

	"github.com/pastdev/clconf/v3/pkg/yamljson"
	"github.com/spf13/cobra"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	yv3 "gopkg.in/yaml.v3"
)

type jsonpathContext struct {
	*rootContext
	first bool
	Marshaler
}

func (c *jsonpathContext) addFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(
		&c.first,
		"first",
		"0",
		false,
		"Prints the first element in the result array")
}

func (c jsonpathContext) jsonpath(
	cmd *cobra.Command, //nolint:unparam
	args []string,
) error {
	path := "$"
	if len(args) > 0 {
		path = args[0]
	}

	data, err := c.rootContext.getValue("/")
	if err != nil {
		return err
	}

	value, err := evaluateJSONPath(path, data, c.first)
	if err != nil {
		return err
	}

	marshalled, err := c.Marshal(value)
	if err != nil {
		return err
	}

	fmt.Print(marshalled)
	return nil
}

func evaluateJSONPath(path string, data interface{}, first bool) (interface{}, error) {
	p, err := yamlpath.NewPath(path)
	if err != nil {
		return nil, fmt.Errorf("invalid jsonpath [%s]: %w", path, err)
	}

	yml, err := yv3.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshalling via yaml.v3: %w", err)
	}
	var n yv3.Node

	err = yv3.Unmarshal(yml, &n)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling to yaml.v3.Node: %w", err)
	}

	results, err := p.Find(&n)
	if err != nil {
		return nil, fmt.Errorf("executing jsonpath.Find: %w", err)
	}

	if first {
		yml, err = yv3.Marshal(results[0])
	} else {
		yml, err = yv3.Marshal(results)
	}
	if err != nil {
		return nil, fmt.Errorf("converting back to yaml via yaml.v3: %w", err)
	}

	result, err := yamljson.UnmarshalSingleYaml(string(yml))
	if err != nil {
		return nil, fmt.Errorf("converting back to clconf interface: %w", err)
	}

	return result, nil
}

func jsonpathCmd(rootCmdContext *rootContext) *cobra.Command {
	var cmdContext = &jsonpathContext{
		rootContext: rootCmdContext,
		Marshaler: Marshaler{
			secretAgentFactory: rootCmdContext,
		},
	}

	var cmd = &cobra.Command{
		Use:   "jsonpath <jsonpath>",
		Short: "Get the value at the supplied path",
		Example: `
  clconf --pipe jsonpath "$..credentials" <<'EOF'
    foodb:
    host: foo.example.com
    credentials:
      username: foouser
      password: foopass
  EOF

	# Output:
  # - password: foopass
  #   username: foouser
		`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdContext.jsonpath(cmd, args)
		},
	}

	cmdContext.addFlags(cmd)
	cmdContext.Marshaler.AddFlags(cmd)

	return cmd
}
