# Templates

Templates are written in Go's [`text/template`](http://golang.org/pkg/text/template/).

## Flat key/value caveats and considerations

Because the [templates only see a flat list of key/value pairs](../README.md#getv-templates), certain operations will behave differently than the CLI (notably `getv` itself).
Take the following yaml for example:

```yaml
credentials:
  username: foo
  password: bar
```

Would be seen by inside the templates as:

```yaml
/credentials/username: foo
/credentials/password: bar
```

In a template, `getv` (and similar value-based functions) can only "see" full keys (e.g. `/credentials/username`).
Asking a template for a partial key (e.g. `/credentials`) will fail.
Additional functions, like `ls` and `lsdir` can provide access to inspecting and ranging on sub-keys.

## Wildcards

Some commands allow for wildcard key matching using `*`.
A `*` does not match `/`.
Multiple `*` are allowed such as `/foo/*/bar/*`.

## Template Functions

### add

Adds two int values.

```console
$ clconf getv / --output go-template --template '{{add 1 3}}'
4
```

### asJson

Converts the supplied value to properly encoded JSON.

```console
$ clconf --pipe getv /foo --output go-template --template '{{asJson (getksvs "/*" "int")}}' <<EOF
foo:
- hip
- hop
- bar
- baz
EOF
["hip","hop","bar","baz"]
```

It's worth noting in this example that we use [`getksvs "/*" "int"`](#getksvs) to extract the values sorted by the `int` value of the keys.
If we didn't do this the array would be in random order because [the data backing getv templates](../README.md#getv-templates) is represented as a map-backed key/value store and map iteration in go is random.

### asJsonString

Converts the supplied value to a properly encoded JSON string.
For any value that is not currently of type string, `fmt.Sprintf("%v", value)` will be used to convert prior to encoding.

```console
$ clconf --pipe getv / --output go-template --template '{{printf "{\"safeforjson\": %s}" (asJsonString (getv "/notsafeforjson"))}}' <<EOF
notsafeforjson: |
  {"']
EOF
{"safeforjson": "{\"']\n"}
```

### atoi

Alias for the [strconv.Atoi](https://golang.org/pkg/strconv/#Atoi) function.

```console
$ clconf getv / --output go-template --template '{{seq 1 (atoi "10")}}'
[1 2 3 4 5 6 7 8 9 10]
```

### base

Alias for the [path.Base](https://golang.org/pkg/path/#Base) function.

```console
$ clconf getv / --output go-template --template '{{base "/foo/bar.txt"}}'
bar.txt
```

### base64Decode

Returns the string representing the decoded base64 value.

```console
$ clconf getv / --output go-template --template '{{base64Decode "VmFsdWU="}}'
Value
```

### base64Encode

Returns a base64 encoded string of the value.

```console
$ clconf getv / --output go-template --template '{{base64Encode "Value"}}'
VmFsdWU=
```

### cget

Equivalent to [`get`](#get) but the value will be decrypted.

### cgets

Equivalent to [`gets`](#gets) but the value will be decrypted.

### cgetv

Equivalent to [`getv`](#getv) but the value will be decrypted.

### cgetvs

Equivalent to [`getvs`](#getvs) but the value will be decrypted.

### contains

Alias for [strings.Contains](https://golang.org/pkg/strings/#Contains)

```console
$ clconf getv / --output go-template --template '{{if contains "a long time ago" "time"}}{{"the world makes sense"}}{{end}}'
the world makes sense
```

### datetime

Alias for [time.Now](https://golang.org/pkg/time/#Now)

```console
clconf getv / --output go-template --template '{{datetime}}'
2023-03-24 10:25:20.282129609 -0600 MDT m=+0.000769001
```

```console
$ clconf getv / --output go-template --template '{{datetime.Format "Jan 2, 2006 at 3:04pm (MST)"}}'
Mar 24, 2023 at 10:25am (MDT)
```

See the time package for more usage: [http://golang.org/pkg/time/](http://golang.org/pkg/time/)

### dir

Equivalent to `path.Dir`

```console
$ clconf getv / --output go-template --template '{{dir "."}}'
.

$ clconf getv / --output go-template --template '{{dir "/foo/bar/bip.txt"}}'
/foo/bar

$ clconf getv / --output go-template --template '{{dir "/foo"}}'
/

$ clconf getv / --output go-template --template '{{dir ""}}'
.
```

### div

Divides two int values.

```console
$ clconf getv / --output go-template --template '{{div 4 2}}'
2

$ clconf getv / --output go-template --template '{{div 3 2}}'
1
```

### escapeOsgi

Places a single `\` prior to any `'`, `"`, `\`, `=` or space.

```console
$ clconf --pipe getv --output go-template --template '{{escapeOsgi "foo=bar"}}' < /dev/null
foo\=bar
```

### exists

Checks if the key exists. Return false if key is not found.

```console
$ clconf --pipe getv / --output go-template --template '{{if exists "/foo/bar"}}exists{{else}}nope{{end}}' <<EOF
foo:
  bar: bip
EOF
exists
```

[Caveat's apply](#flat-keyvalue-caveats-and-considerations):

```console
$ clconf --pipe getv / --output go-template --template '{{if exists "/foo"}}exists{{else}}nope{{end}}' <<EOF
foo:
  bar: bip
EOF
nope
```

### fileExists

Checks if the file or directory at the specified filesystem path exists.

```text
{{if fileExists "/etc/myConfig"}}
    useConfig: /etc/myConfig
{{end}}
```

### fqdn

Adds a domain to a hostname if not already qualified.

```console
$ clconf --pipe getv --output go-template --template '{{fqdn "foo" "example.com"}}' < /dev/null
foo.example.com
$ clconf --pipe getv --output go-template --template '{{fqdn "foo.google.com" "example.com"}}' < /dev/null
foo.google.com
```

### get

Returns the KVPair where key matches its argument.
Returns an error if key is not found.
Wildcards not supported.

```console
$ clconf --pipe getv / --output go-template --template '{{with get "/foo/bar"}}k: {{.Key}}, v: {{.Value}}{{end}}' <<EOF
foo:
  bar: bip
EOF

$ clconf --pipe getv / --output go-template --template '{{with get "/foo/*"}}k: {{.Key}}, v: {{.Value}}{{end}}' <<EOF
foo:
  bar: bip
EOF
Error: template execute: execute template: template: cli:1:7: executing "cli" at <get "/foo/*">: error calling get: /foo/*: key does not exist
```

[Caveat's apply](#flat-keyvalue-caveats-and-considerations):

```console
$ clconf --pipe getv / --output go-template --template '{{with get "/foo"}}k: {{.Key}}, v: {{.Value}}{{end}}' <<EOF
foo:
  bar: bip
EOF
Error: template execute: execute template: template: cli:1:7: executing "cli" at <get "/foo">: error calling get: /foo: key does not exist
```

### getenv

Wrapper for [os.Getenv](https://golang.org/pkg/os/#Getenv). Retrieves the value of the environment variable named by the key. It returns the value, which will be empty if the variable is not present. Optionally, you can give a default value that will be returned if the key is not present.

```console
MYENV=foo clconf getv / --output go-template --template '[{{getenv "MYENV"}}]'
[foo]

$ MYENV= clconf getv / --output go-template --template '[{{getenv "MYENV"}}]'
[]

# Default used when defined and empty
$ MYENV= clconf getv / --output go-template --template '[{{getenv "MYENV" "adefault"}}]'
[adefault]

$ clconf getv / --output go-template --template '[{{getenv "MYENV" "adefault"}}]'
[adefault]

# Does not fail if not defined
$ clconf getv / --output go-template --template '[{{getenv "MYENV"}}]'
[]
```

### getksvs

Returns all values, []string, where key matches its argument, sorted by key.
Specify optional argument `int` to sort the keys as integers.
Returns an error if key is not found.
Wildcards are allowed.

Preserve the order of list inputs:

```console
$ clconf --pipe getv --output go-template --template '{{getksvs "/foo/*" "int"}}' <<EOF
foo:
- dog
- bird
- cat
EOF
[dog bird cat]
```

Sort by string keys:

```console
$ clconf --pipe getv --output go-template --template '{{getksvs "/foo/*"}}' <<EOF
foo:
  dog: woof
  bird: tweet
  cat: meow
EOF
[tweet meow woof]

$ clconf --pipe getv --output go-template --template '{{getksvs "/*/*"}}' <<EOF
foo:
  dog: woof
  bird: tweet
  cat: meow
bar:
  fox: ????
EOF
[tweet meow woof ????]

$ clconf --pipe getv --output go-template --template '{{getksvs "/foo/dog"}}' <<EOF
foo:
  dog: woof
  bird: tweet
  cat: meow
EOF
[woof]
```

[Caveat's](#flat-keyvalue-caveats-and-considerations) and [wildcard rules](#wildcards) apply:

```console
$ clconf --pipe getv --output go-template --template '{{getksvs "/*"}}' <<EOF
foo:
  dog: woof
  bird: tweet
  cat: meow
EOF
[]

$ clconf --pipe getv --output go-template --template '{{getksvs "/foo"}}' <<EOF
foo:
  dog: woof
  bird: tweet
  cat: meow
EOF
[]
```

### getsvs

Returns all values, []string, where key matches its argument, sorted.
Optionally specify `int` to sort the values as integers. Returns an error if key is not found.
Wildcards optional.

```console
$ clconf --pipe getv --output go-template --template '{{getsvs "/foo/*"}}' <<EOF
foo:
- dog
- bird
- cat
EOF
[bird cat dog]

$ clconf --pipe getv --output go-template --template '{{getsvs "/foo/*"}}' <<EOF
foo:
  dog: woof
  bird: tweet
  cat: meow
EOF
[meow tweet woof]

$ clconf --pipe getv --output go-template --template '{{getsvs "/foo/dog"}}' <<EOF
foo:
  dog: woof
  bird: tweet
  cat: meow
EOF
[woof]

$ clconf --pipe getv --output go-template --template '{{getsvs "/*/*"}}' <<EOF
foo:
  dog: woof
  bird: tweet
  cat: meow
bar:
  fox: ????
EOF
[???? meow tweet woof]
```

[Caveat's](#flat-keyvalue-caveats-and-considerations) and [wildcard rules](#wildcards) apply:

```console
$ clconf --pipe getv --output go-template --template '{{getsvs "/*"}}' <<EOF
foo:
  dog: woof
  bird: tweet
  cat: meow
EOF
[]

$ clconf --pipe getv --output go-template --template '{{getsvs "/foo"}}' <<EOF
foo:
  dog: woof
  bird: tweet
  cat: meow
EOF
[]
```

### gets

Returns all KVPair, []KVPair, where key matches its argument.
Returns an error if key is not found.
Wildcards optional.

```console
$ clconf --pipe getv / --output go-template --template '{{range gets "/foo/*"}}k: {{.Key}}, v: {{.Value}}{{"\n"}}{{end}}' <<EOF
foo:
  bar: bip
  zip: zap
EOF
k: /foo/bar, v: bip
k: /foo/zip, v: zap

$ clconf --pipe getv / --output go-template --template '{{range gets "/foo/bar"}}k: {{.Key}}, v: {{.Value}}{{"\n"}}{{end}}' <<EOF
foo:
  bar: bip
  zip: zap
EOF
k: /foo/bar, v: bip

$ clconf --pipe getv / --output go-template --template '{{range gets "/*/*"}}k: {{.Key}}, v: {{.Value}}{{"\n"}}{{end}}' <<EOF
foo:
  bar: bip
  zip: zap
EOF
k: /foo/bar, v: bip
k: /foo/zip, v: zap
```

[Caveat's](#flat-keyvalue-caveats-and-considerations) and [wildcard rules](#wildcards) apply:

```console
$ clconf --pipe getv / --output go-template --template '{{range gets "/*"}}k: {{.Key}}, v: {{.Value}}{{"\n"}}{{end}}' <<EOF
foo:
  bar: bip
  zip: zap
EOF

$ clconf --pipe getv / --output go-template --template '{{range gets "/foo"}}k: {{.Key}}, v: {{.Value}}{{"\n"}}{{end}}' <<EOF
foo:
  bar: bip
  zip: zap
EOF
```

### getv

Returns the value as a string where key matches its argument or an optional default value.
Returns an error if key is not found and no default value given.
Wildcards not supported.

```console
$ clconf --pipe getv / --output go-template --template '{{getv "/foo/bar"}}' <<EOF
foo:
  bar: bip
  zip: zap
EOF
bip

# With default value
$ clconf --pipe getv / --output go-template --template '{{getv "/foo/hip" "hop"}}' <<EOF
foo:
  bar: bip
  zip: zap
EOF
hop

# Missing, no default
$ clconf --pipe getv / --output go-template --template '{{getv "/foo/hip"}}' <<EOF
foo:
  bar: bip
  zip: zap
EOF
Error: template execute: execute template: template: cli:1:2: executing "cli" at <getv "/foo/hip">: error calling getv: /foo/hip: key does not exist
```

[Caveat's apply](#flat-keyvalue-caveats-and-considerations):

```console
$ clconf --pipe getv / --output go-template --template '{{getv "/foo"}}' <<EOF
foo:
  bar: bip
  zip: zap
EOF
Error: template execute: execute template: template: cli:1:2: executing "cli" at <getv "/foo">: error calling getv: /foo: key does not exist
```

### getvs

Returns all values, []string, where key matches its argument, string-sorted.
Wildcard required.

```console
$ clconf --pipe getv / --output go-template --template '{{getvs "/foo/*"}}' <<EOF
foo:
  buzz: 1 
  zap: 10
  hop: 2
EOF
[1 10 2]

$ clconf --pipe getv / --output go-template --template '{{getvs "/bar/*"}}' <<EOF
foo:
  buzz: 1 
  zap: 10
  hop: 2
EOF
[]

clconf --pipe getv / --output go-template --template '{{getvs "/foo/buzz"}}' <<EOF
foo:
  buzz: 1 
  zap: 10
  hop: 2
EOF
[]

# Multiple wildcards supported
$ clconf --pipe getv / --output go-template --template '{{getvs "/*/*"}}' <<EOF
foo:
  buzz: 1 
  zap: 10
  hop: 2
bar:
  hip: hop
  eleven: 11
EOF
[1 10 11 2 hop]
```

### join

Alias for the [strings.Join](https://golang.org/pkg/strings/#Join) function.

```console
$ clconf --pipe getv / --output go-template --template '{{join (getvs "/things/*") ","}}' <<EOF
things:
- thing1
- thing2
EOF
thing1,thing2
```

### json

Returns an map[string]interface{} of the json value.

```console
$ clconf --pipe getv / --output go-template --template '{{(json (getv "/json_obj_string")).foo}}' <<EOF
json_obj_string: |
  {"foo": "bar"}
EOF
bar
```

### jsonArray

Returns a []interface{} from a json array such as `["a", "b", "c"]`.

```console
$ clconf --pipe getv / --output go-template --template '{{range (jsonArray (getv "/json_array_string"))}}{{ . }}{{"\n"}}{{end}}' <<EOF
json_array_string: |
  ["foo", "bar"]
EOF
foo
bar
```

### lookupIP

Wrapper for [net.LookupIP](https://golang.org/pkg/net/#LookupIP) function.
The wrapper also sorts (alphabeticaly) the IP addresses.
This is crucial since in dynamic environments DNS servers typically shuffle the addresses linked to domain name.
And that would cause unnecessary config reloads.

```text
$ clconf getv / --output go-template --template '{{lookupIP "localhost"}}'
[127.0.0.1]
```

### lookupIPV4

Same as `lookupIP` but filters down to IPV4 adresses.

### lookupIPV6

Same as `lookupIP` but filters down to IPV6 adresses.

### lookupSRV

Wrapper for [net.LookupSRV](https://golang.org/pkg/net/#LookupSRV).
The wrapper also sorts the SRV records alphabetically by combining all the fields of the net.SRV struct to reduce unnecessary config reloads.

```text
{{range lookupSRV "mail" "tcp" "example.com"}}
  target: {{.Target}}
  port: {{.Port}}
  priority: {{.Priority}}
  weight: {{.Weight}}
{{end}}
```

### ls

If the search string exactly matches a key, returns the equivalent of `base`.
If the search string exists and has subkeys, returns a list of the subkeys.
Returns an empty list if search string is not found.
Wildcards are not supported.

```console
$ clconf --pipe getv / --output go-template --template '[{{join (ls "/foo") ","}}]' <<EOF
foo:
  bar: bip
  hip:
    hop: hoop
  zip: zap
EOF
[bar,hip,zip]

$ clconf --pipe getv / --output go-template --template '[{{join (ls "/foo/hip") ","}}]' <<EOF
foo:
  bar: bip
  hip:
    hop: hoop
  zip: zap
EOF
[hop]

$ clconf --pipe getv / --output go-template --template '[{{join (ls "/foo/hip/hop") ","}}]' <<EOF
foo:
  bar: bip
  hip:
    hop: hoop
  zip: zap
EOF
[hop]

$ clconf --pipe getv / --output go-template --template '[{{join (ls "/") ","}}]' <<EOF
foo:
  bar: bip
  hip:
    hop: hoop
  zip: zap
EOF
[foo]

$ clconf --pipe getv / --output go-template --template '[{{join (ls "/bar") ","}}]' <<EOF
foo:
  bar: bip
  hip:
    hop: hoop
  zip: zap
EOF
[]
```

### lsdir

Returns all subkeys that are not full keys when appended to the search string.

```console
# Note 'bar' is missing from the returned value because /foo/bar is a full key
$ clconf --pipe getv / --output go-template --template '[{{join (lsdir "/foo") ","}}]' <<EOF
foo:
  bar: baz
  hip:
  - hop: hap
  - hup: hep
  zip:
    zap: bop
EOF
[hip,zip]

$ clconf --pipe getv / --output go-template --template '[{{join (lsdir "/foo/hip") ","}}]' <<EOF
foo:
  bar: baz
  hip:
  - hop: hap
  - hup: hep
  zip:
    zap: bop
EOF
[0,1]

# No results here because /foo/zip/zap would be a full key
$ clconf --pipe getv / --output go-template --template '[{{join (lsdir "/foo/zip") ","}}]' <<EOF
foo:
  bar: baz
  hip:
  - hop: hap
  - hup: hep
  zip:
    zap: bop
EOF
[]
```

### map

Creates a key-value map of string -> interface{}

```text
{{$endpoint := map "name" "elasticsearch" "private_port" 9200 "public_port" 443}}

name: {{index $endpoint "name"}}
private-port: {{index $endpoint "private_port"}}
public-port: {{index $endpoint "public_port"}}
```

specifically useful if you use a sub-template and you want to pass multiple values to it.

### mod

Modulus of two int values.

```text
$ clconf getv / --output go-template --template '{{mod 10 3}}'
1
```

### parseBool

An alias to [`strconv.ParseBool`](https://golang.org/pkg/strconv/#ParseBool)

```console
$ clconf getv / --output go-template --template '{{parseBool "true"}}'
true

$ clconf getv / --output go-template --template '{{parseBool "T"}}'
true

$ clconf getv / --output go-template --template '{{parseBool "F"}}'
false

$ clconf getv / --output go-template --template '{{parseBool "1"}}'
true

$ clconf getv / --output go-template --template '{{parseBool "R"}}'
Error: template execute: execute template: template: cli:1:2: executing "cli" at <parseBool "R">: error calling parseBool: strconv.ParseBool: parsing "R": invalid syntax

$ clconf getv / --output go-template --template '{{parseBool "-1"}}'
Error: template execute: execute template: template: cli:1:2: executing "cli" at <parseBool "-1">: error calling parseBool: strconv.ParseBool: parsing "-1": invalid syntax

$ clconf getv / --output go-template --template '{{parseBool "0"}}'
false
```

### regexReplace

Given a regex, an original string and a replacement string run [`regexp.ReplaceAllString`](https://golang.org/pkg/regexp/#Regexp.ReplaceAllString) and return the result. Returns an error if the regex fails to compile.

```console
$ clconf --pipe getv --output go-template --template '{{regexReplace "o+" "foo" "e"}}' < /dev/null
fe
```

### replace

Alias for the [strings.Replace](https://golang.org/pkg/strings/#Replace) function.

```console
$ clconf getv / --output go-template --template '{{replace "foo" "o" "e" -1}}'
fee

$ clconf getv / --output go-template --template '{{replace "foo" "o" "e" 1}}'
feo
```

### reverse

Reverses a list.  If the list is `KVPair`, will compare keys, not values.

### seq

Alias for the [template.Seq](https://golang.org/pkg/template/#Seq) function.

```console
$ clconf getv / --output go-template --template '{{range (seq 1 10) }}{{.}}{{"\n"}}{{end}}'
1
2
3
4
5
6
7
8
9
10
```

### sort

Sorts the input ([]interface{}) by translating it to the specified type (one of `int`, `string`, default: `string`)

```console
clconf --pipe getv --output go-template --template '{{sort (getvs "/foo/*") }}' <<EOF
foo:
- dog
- bird
- cat
EOF
[bird cat dog]

clconf --pipe getv --output go-template --template '{{ range $i := (sort (ls "/foo") "int")}}{{ getv (printf "/foo/%s" $i) }},{{ end }}' <<EOF
foo:
- dog
- bird
- cat
EOF
dog,bird,cat,
```

### sortByLength

Sorts a list of `string` values by their length.

### sortKvByLength

Sorts a list of `KVPair` values by the length of their key.

### split

Alias for [strings.Split](http://golang.org/pkg/strings/#Split). Splits the input string on the separating string and returns a slice of substrings.

```text
{{ $url := split (getv "/deis/service") ":" }}
    host: {{index $url 0}}
    port: {{index $url 1}}
```

### toLower

Alias for [strings.ToLower](http://golang.org/pkg/strings/#ToLower). Returns lowercased string.

```text
key: {{toLower "Value"}}
```

### toUpper

Alias for [strings.ToUpper](http://golang.org/pkg/strings/#ToUpper). Returns uppercased string.

```text
key: {{toUpper "value"}}
```

### trimSuffix

Alias for [strings.TrimSuffix](http://golang.org/pkg/strings/#TrimSuffix).

## Example Usage

Given the yaml input:

```yaml
---
nginx:
  domain: example.com
  root: /var/www/example_dotcom
  worker_processes: 2
app:
  upstream:
    app1: 10.0.1.100:80
    app2: 10.0.1.101:80
```

And the template:

```text
worker_processes {{getv "/nginx/worker_processes"}};

upstream app {
{{range getvs "/app/upstream/*"}}
    server {{.}};
{{end}}
}

server {
    listen 80;
    server_name www.{{getv "/nginx/domain"}};
    access_log /var/log/nginx/{{getv "/nginx/domain"}}.access.log;
    error_log /var/log/nginx/{{getv "/nginx/domain"}}.log;

    location / {
        root              {{getv "/nginx/root"}};
        index             index.html index.htm;
        proxy_pass        http://app;
        proxy_redirect    off;
        proxy_set_header  Host             $host;
        proxy_set_header  X-Real-IP        $remote_addr;
        proxy_set_header  X-Forwarded-For  $proxy_add_x_forwarded_for;
    }
}
```

Output:

```text
worker_processes 2;

upstream app {
    server 10.0.1.100:80;
    server 10.0.1.101:80;
}

server {
    listen 80;
    server_name www.example.com;
    access_log /var/log/nginx/example.com.access.log;
    error_log /var/log/nginx/example.com.error.log;

    location / {
        root              /var/www/example_dotcom;
        index             index.html index.htm;
        proxy_pass        http://app;
        proxy_redirect    off;
        proxy_set_header  Host             $host;
        proxy_set_header  X-Real-IP        $remote_addr;
        proxy_set_header  X-Forwarded-For  $proxy_add_x_forwarded_for;
    }
}
```

### Complex example

This examples show how to use a combination of the templates functions to do nested iteration.

```text
{{range $dir := lsdir "/services/web"}}
upstream {{base $dir}} {
    {{$custdir := printf "/services/web/%s/*" $dir}}{{range gets $custdir}}
    server {{$data := json .Value}}{{$data.IP}}:80;
    {{end}}
}

server {
    server_name {{base $dir}}.example.com;
    location / {
        proxy_pass {{base $dir}};
    }
}
{{end}}
```

Output:

```text
upstream cust1 {
    server 10.0.0.1:80;
    server 10.0.0.2:80;
}

server {
    server_name cust1.example.com;
    location / {
        proxy_pass cust1;
    }
}

upstream cust2 {
    server 10.0.0.3:80;
    server 10.0.0.4:80;
}

server {
    server_name cust2.example.com;
    location / {
        proxy_pass cust2;
    }
}
```

Go's [`text/template`](http://golang.org/pkg/text/template/) package is very powerful. For more details on it's capabilities see its [documentation.](http://golang.org/pkg/text/template/)
