# clconf

`clconf` provides a utility for merging multiple config files and extracting
values using a path string.  `clconf` is both a _library_, and a _command line
application_.

For details, see `clconf --help`.

## Background

`clconf` was primarily designed for use in
containers where you inject secrets as files/environment variables, but
need to convert them to application specific configuration files.

The [12 factor app](https://12factor.net/config) states:

> The twelve-factor app stores config in environment variables

But many existing applications/frameworks expect configuration files.
This application helps serve to bridge this gap.  `clconf` can be used
by itself, or combined with other tools like
[confd](https://github.com/pastdev/confd) (_note, this is my fork of
confd as
[they chose a different direction](https://github.com/kelseyhightower/confd/pull/663)_).

## Configuration

Using `clconf` requires one or more yaml files (or strings) to merge together.
They are specified using either using environment variables or command line
options, as either files or base64 encoded strings.  The order they are
processed in is as follows:

1. _`--yaml`_: One or more files.
1. _`YAML_FILES` environment variable_: A comma separated list of files.
1. _`--yaml-base64`_: One or more base64 encoded strings containing yaml.
1. _`YAML_VARS` environment variable_: A comma separated list of environment
  variable names, each a base64 encoded string containing yaml.
1. _`--stdin`_: One or more `---` separated yaml files read from `stdin`.
1. _`--var`_: One or more path overrides of the form `/foo="bar"`.  Key is a
  path, an value is json/yaml encoded.
1. _`--patch`_: One or more rfc 6902 json/yaml patch files to apply to the
  result of merging all the config sources.
1. _`--patch-string`_: One or more rfc 6902 json/yaml patches to apply to the
  result of merging all the config sources.

All of these categories of input will be appended to each other and the _last
defined value of any key will take precedence_.  For example:

```bash
YAML_FILES="c.yml,d.yml"
YAML_VARS="G_YML_B64,H_YML_B64"
E_YML_B64"$(echo -e "c:\n  foo: bar" | base64 -w 0)
F_YML_B64"$(echo -e "d:\n  foo: bar" | base64 -w 0)

G_YML_B64="$(echo -e "g:\n  foo: bar" | base64 -w 0)
H_YML_B64="$(echo -e "h:\n  foo: bar" | base64 -w 0)

clconf \
  --yaml a.yml \
  --yaml b.yml \
  --yaml-base64 "$E_YML_B64" \
  --yaml-base64 "$F_YML_B64" \
  --var '/foo="bar"' \
  --patch patch.json \
  --patch-string '[{"op": "replace", "path": "/foo", "value": "baz"}]' \
  <<<"---\nfoo: baz"
```

Would be processed in the following order:

1. `a.yml`
1. `b.yml`
1. `c.yml`
1. `d.yml`
1. `E_YML_B64`
1. `F_YML_B64`
1. `G_YML_B64`
1. `H_YML_B64`
1. `stdin`
1. `/foo="bar"`
1. `patch.json`
1. `[{"op": "replace", "path": "/foo", "value": "baz"}]`

## Use Cases

### Helper in Scripts

#### Getv as JSON

When using the `--as-json` option, the value obtained at the indicated path
will be serialized to json.  For example, if you have `foo.yml`:

```yaml
applications:
- a
- b
- c
```

You could use:

```bash
clconf --yaml foo.yml getv /applications --as-json
```

To get:

```json
["a","b","c"]
```

#### Convert Bash Array to JSON

```bash
arr=(a b)

# always use the `--` to ensure none of the arguments get consumed by clconf
clconf var -- /foo "${arr[@]}" # /foo=["a","b"]

# note that with --value-only the / is ignored and can be anything
clconf --value-only -- / "${arr[@]}" # ["a","b"]

# can force single values into array
clconf --force-array -- /root/arr "foo" # /root/arr=["foo"]
```

#### Convert JSON Array to Bash

Allows for iteration:

```bash
# clconf getv --as-bash array will print out '([0]="foo bar" [1]="hip hop")'
# which bash's declare -a can turn into an array.  using --var here for
# simplicity but any yaml/json source will do.
declare -a arr="$(clconf --var '/a=["foo bar", "hip hop"]' getv /a --as-bash-array)"
for i in "${arr[@]}"; do
  printf '<<<%s>>>' "$i"
done # prints out <<<foo bar>>><<<hip hop>>>
```

Also allows for iteration over maps:

```bash
declare -a arr="$(clconf --var '/a={"foo": "bar","hip":"hop"}' getv /a --as-bash-array)"
for i in "${arr[@]}"; do
  printf '<<<%s>>>' "$i"
done # <<<{"key":"foo","value":"bar"}>>><<<{"key":"hip","value":"hop"}>>>
```

#### Get Value Using JSON Path

The `jsonpath` subcommand allows you to use jsonpath syntax to locate values.
The values obtained have the same output formatting options as `getv` does.

```bash
clconf --pipe jsonpath "$..credentials" <<'EOF'
foodb:
  host: foo.example.com
  credentials:
    username: foouser
    password: foopass
EOF

# - password: foopass
#   username: foouser
```

#### Getv Templates

Templates allow you to apply your configuration to golang template plus some
additional [custom functions](docs/templates.md)

Note that when used in conjunction with the `--template` options,
`getv` templates see a one-level key-value map, not the map
represented by the yaml.  For example, this yaml (`foo.yml`):

```yaml
applications:
- a
- b
- c
credentials:
  username: foo
  password: bar
```

Would be seen by inside the templates as:

```yaml
/applications/0: a
/applications/1: b
/applications/2: c
/credentials/username: foo
/credentials/password: bar
```

A simple bash program to utilize this might look like:

```bash
#!/bin/bash

set -e

function applications {
  run_clconf \
    getv '/' \
    --template-string "
      {{- range getvs \"/applications/*\" }}
        {{- . }}
      {{ end }}"
}

function getv {
  local path=$1
  run_clconf getv "${path}"
}

function run_clconf {
  ./clconf --ignore-env --yaml 'foo.yml' "$@"
}

user="$(getv "/credentials/username")"
pass="$(getv "/credentials/password")"
applications | xargs -I {} {} --user "${user}" --pass "${pass}"
```

### Kubernetes/OpenShift

This is my primary use case.  It is a natural extension of the
built-in `ConfigMap` and `Secret` objects.  With `clconf` you can
provide non-sensitive environment configuration in a `ConfigMap`:

```yaml
db:
  url: jdbc.mysql:localhost:3306/mydb
```

and sensitive configuration in a `Secret`:

```yaml
db:
  username: mydbuser
  password: youllneverguess
```

Then with:

```bash
clconf \
  --yaml /etc/myapp/config.yml \
  --yaml /etc/myapp/secrets.yml \
  getv \
  > /app/config/application.yml
```

You would have a file containing:

```yaml
db:
  url: jdbc.mysql:localhost:3306/mydb
  username: mydbuser
  password: youllneverguess
```

Which can be written to an in-memory `emptyDir`:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: app
spec:
  containers:
    - name: app
      image: my-springboot-app
      volumeMounts:
        - name: app-config
          mountPath: /app/config
  volumes:
  - name: app-config
    emptyDir:
      medium: "Memory"
```

So that the sensitive information never touches disk and would not be exposed by
a `ps` command.

### Secret Management

`clconf` can encrypt and decrypt values as well, similar in nature
to `ansible-vault`.  This allows you to commit your _secrets_ alongside
the code that uses them.  For example, you could create a new config
file:

```yaml
db:
  url: jdbc.mysql:localhost:3306/mydb
```

Then add your secrets:

```bash
clconf \
  --secret-keyring testdata/test.secring.gpg \
  --yaml C:/Temp/config.yml \
  csetv /db/username dbuser
clconf \
  --secret-keyring testdata/test.secring.gpg \
  --yaml C:/Temp/config.yml \
  csetv /db/password dbpass
```

Which would result in something safe to commit with your source code:

```yaml
db:
  password: wcBMA5B5A4w5Zw+rAQgAJ9bR77oJi0P7X5qtnN+soUCszYTy6VGvNutHInE0QCugyXhVeovm+iPaFo/K5D8IO9QJnRL4D9PCiuqVslhsP54b7Qpep/1R/1HEbw9XNMv+uTh9CQDnT1FMer9i+samZ6poTT5uWMJtdTnwa187V5TUGKQdSwoz82CgQ8zQYq0aI15kZp4VziN9eQV1jrphG2+aJdtyIuIouafuEMSnrRz+bb8xAWu3I1INfEP0MuttTYdoY9W3xEU7L4IGvzhw8rnJPNhkK5LKTtvlOCDpKSs1ESReBHYSPNSAAlKBOTHwZ1MHKnypiWVzGACzq+Yh0K+UGtb8dGRiFhwMAn9jfdLgAeRkS/i2wGBjd3suaPzadgW84a0e4L3g3+FKo+Co4k3c3CHgB+OodVAQ2+LoReD54e6X4HbjH52aGIGkSKbg5eLA4qGv4Dnjwf422VOoqubgTOQV3gjv0NTKLF9IXaFPyhtj4joDyk/hwo8A
  url: jdbc.mysql:localhost:3306/mydb
  username: wcBMA5B5A4w5Zw+rAQgAAH1FM4x/FAjmspKbyHJvvaMwmFjGOMOKIle1oe0tpewzaUaEoYZ2trx8nerbWqtIxf4rnB9kNA2YyKs6CLka1q6jnN2U4KI3EjXQaaf6sL5qg/g3Hlak937Wf8+fK1tpghGuFJXTcRjqOgAyV8LfZtQ7MDfgoIy30bihjQz/0TzNi3IZlezqsgvLqoRsgP4b5S9liR/8EaQQ9BepaAgjl3c37QJf/qQK1mkPTOGzlTzZ7dcicpycxRwU8mMlYMq4qN0RR8ZMuiPshYJOdb3OVbNZq08MVzRbuMcPo+SbJsckD+V7EvOn3Km7jefblZsx2fzRPrAG23zZYkAPsUUuE9LgAeTO9rtOh0NQhkYL+9nJzCE+4dpv4K3gCOGkxOBR4o5q737gIuOVjW3r5vC/cuCA4ciT4JDjUV+uW8+IzSfgceKckR304HrjfbEkfn2gljvgAuSCU2yJMaO1aVjs225Rhw7q4pq3xL3hDV4A
```

These values can be decrypted using:

```bash
clconf \
  --secret-keyring testdata/test.secring.gpg \
  --yaml C:/Temp/config.yml \
  cgetv /db/username
clconf \
  --secret-keyring testdata/test.secring.gpg \
  --yaml C:/Temp/config.yml \
  cgetv /db/password
```

Or in conjunction with templates

```bash
clconf \
  --secret-keyring testdata/test.secring.gpg \
  --yaml C:/Temp/config.yml \
  getv / \
  --template-string '{{ cgetv "/db/username" }}:{{ cgetv "/db/password" }}'
```

### Templating

`clconf` has a `template` operation that functions as a
[confd](https://github.com/kelseyhightower/confd) replacement but supports only
yaml as a value store. It uses command line arguments in place of `confd`'s
[toml files](https://github.com/kelseyhightower/confd/blob/master/docs/template-resources.md)
to determine where templates are found and output placed.

`clconf` supports [additional functions](templates.md) above what `confd`
provides.

All of the options for `getv` are available for specifying yaml sources,
and the templates behave as outlined above. The `template` operation takes it
a step further by templating many files in a single run. The `template`
function's `--help` provides examples.

See the [template function documentation](docs/templates.md) for the available
template functions.
