# clconf
`clconf` provides a utility for merging multiple config files and extracting
values using a path string.  

See `clconf --help` for details.

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

## Use Cases

### Kubernetes/OpenShift
This is my primary use case.  It is a natural extension of the
built-in `ConfigMap` and `Secret` objects.  With `clconf` you can
provide non-sensitive environment configuration in a `ConfigMap`:
```
db: 
  url: jdbc.mysql:localhost:3306/mydb
```
and sensitive configuration in a `Secret`:
```
db: 
  username: mydbuser
  password: youllneverguess
```
Then with:
```
clconf getv --yaml /etc/myapp/config.yml --yaml /etc/myapp/secrets.yml
```
You would get:
```
db: 
  url: jdbc.mysql:localhost:3306/mydb
  username: mydbuser
  password: youllneverguess
```