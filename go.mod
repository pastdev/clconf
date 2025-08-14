module github.com/pastdev/clconf/v3

require (
	dario.cat/mergo v1.0.2
	github.com/evanphx/json-patch/v5 v5.9.11
	github.com/mitchellh/mapstructure v1.5.0
	github.com/spf13/cobra v1.9.1
	github.com/stretchr/testify v1.10.0
	github.com/vmware-labs/yaml-jsonpath v0.3.2
	github.com/xordataexchange/crypt v0.0.3-0.20170626215501-b2862e3d0a77
	// currently locked yaml at these lower levels, we need v2 because v3
	// marshals with all lists indented and no option to change that behavior
	// and v3 locked becuase update higher and the maintainer changed to
	//   https://github.com/yaml/go-yaml
	// with a new package structure:
	//   go.yaml.in/yaml/v[1,2,3,4]
	// which breaks other deps like yaml-jsonpath
	//   https://github.com/vmware-labs/yaml-jsonpath/issues/63
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dprotaso/go-yit v0.0.0-20240618133044-5a0af90af097 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.7 // indirect
	golang.org/x/crypto v0.41.0 // indirect
)

go 1.23.0

toolchain go1.24.6
