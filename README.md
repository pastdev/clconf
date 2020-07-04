# clconf

`clconf` provides a utility for merging multiple config files and extracting values using a path string.  `clconf` is both a _library_, and a _command line application_.

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

Using `clconf` requires one or more yaml files (or strings) to merge together.  They are specified using either using environment variables or command line options, as either files or base64 encoded strings.  The order they are processed in is as follows:

1. *`--yaml`*: One or more files.
1. *`YAML_FILES` environment variable*: A comma separated list of files.
1. *`--yaml-base64`*: One or more base64 encoded strings containing yaml.
1. *`YAML_VARS` environment variable*: A comma separated list of environment variable names, each a base64 encoded string containing yaml.
1. *`--stdin`*: One or more `---` separated yaml files read from `stdin`.
1. *`--var`*: One or more path overrides of the form `/foo="bar"`.  Key is a path, an value is json/yaml encoded.

All of these categories of input will be appended to each other and the _last defined value of any key will take precedence_.  For example:

```bash
YAML_FILES="a.yml,b.yml"
YAML_VARS="C_YML_B64,D_YML_B64"
C_YML_B64"$(echo -e "c:\n  foo: bar" | base64 -w 0)
D_YML_B64"$(echo -e "d:\n  foo: bar" | base64 -w 0)

G_YML_B64="$(echo -e "g:\n  foo: bar" | base64 -w 0)
H_YML_B64="$(echo -e "h:\n  foo: bar" | base64 -w 0)

clconf \
  --yaml e.yml \
  --yaml f.yml \
  --yaml-base64 "$G_YML_B64" \
  --yaml-base64 "$H_YML_B64" \
  --var '/foo="bar"' \
  <<<"---\nfoo: baz"
```

Would be processed in the following order:

1. `a.yml`
1. `b.yml`
1. `C_YML_B64`
1. `D_YML_B64`
1. `e.yml`
1. `f.yml`
1. `G_YML_B64`
1. `H_YML_B64`
1. `stdin`
1. `/foo="bar"`

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
clconf --var "$(clconf var /foo "$@")" getv /foo --as-json
```

#### Getv Templates

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

```yaml
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

All of the options for `getv` are available for specifying yaml sources,
and the templates behave as outlined above. The `template` operation takes it
a step further by templating many files in a single run. The `template`
function's `--help` provides examples:

```bash
This will take an arbitrary number of source templates (or folders full
of templates) and process them either in place (see --in-place) or into the
folder specified as the last argument. It will make any folders required
along the way. If a source is an existing file (not a folder) it will be
treated as a template regardless of the extension (though if the extension
matches it will still be removed).

Usage:
  clconf template <src1> [src2...] [destination folder] [flags]

Examples:

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


Flags:
      --dir-mode string             Chmod mode (e.g. 755) to apply to newly created directories. (default "775")
      --file-mode string            Chmod mode (e.g. 644) to apply to files when templating (new and existing) (defaults to copy from source template).
      --flatten                     Don't preserve relative folders when processing a source folder.
  -h, --help                        help for template
      --in-place                    Template the files in the folder they're found (implies no destination)
      --keep-empty                  Keep empty (zero byte) result files (the default is to remove them)
      --keep-existing-permissions   Only apply --file-mode to new files, leave existing files as-is.
      --rm                          Remove template files after processing.
      --template-extension string   Template file extension (will be removed during templating). (default ".clconf")
```
