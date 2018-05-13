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
clconf \
    --yaml /etc/myapp/config.yml \
    --yaml /etc/myapp/secrets.yml \
    getv \
    > /app/config/application.yml
```
You would have a file containing:
```
db: 
  url: jdbc.mysql:localhost:3306/mydb
  username: mydbuser
  password: youllneverguess
```
Which can be written to an in-memory `emptyDir`:
```
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
So that the sensitive information never touches disk and would not be 
exposed by a `ps` command.

There is also the built-in templating that can be used for more
complicated situations.  For example, you can create a template file
for tomcat's `server.xml`:
```
<?xml version='1.0' encoding='utf-8'?>
<Server port="{{ getv "/shutdown/port" }}" shutdown="{{ getv "/shutdown/password" }}">
  <Listener className="org.apache.catalina.core.JasperListener" />
  <Listener className="org.apache.catalina.core.AprLifecycleListener" SSLEngine="on" />
  <Listener className="org.apache.catalina.core.JreMemoryLeakPreventionListener" />
  <Listener className="org.apache.catalina.mbeans.GlobalResourcesLifecycleListener" />
  <Listener className="org.apache.catalina.core.ThreadLocalLeakPreventionListener" />
  
  <GlobalNamingResources>
    <Resource name="jdbc/appdb"
            auth="Container"
            type="javax.sql.DataSource"
            username="{{ getv "/db/username" }}"
            password="{{ getv "/db/password" }}"
            driverClassName="{{ getv "/db/driver" }}"
            url="{{ getv "/db/url" }}"
            maxActive="8"
            maxIdle="4"/>
  </GlobalNamingResources>
  
  <Service name="Catalina">
    <Connector port="{{ getv "/port" }}" protocol="HTTP/1.1" />
  
    <Engine name="Catalina" defaultHost="localhost">
      <Host name="localhost"
          appBase="webapps"
          unpackWARs="true"
          autoDeploy="true">
        <Valve className="org.apache.catalina.valves.AccessLogValve"
            directory="logs"
            prefix="localhost_access_log."
            suffix=".txt"
            pattern="%h %l %u %t "%r" %s %b" />
      </Host>
    </Engine>
  </Service>
</Server>
```
Then use:
```
clconf \
    --yaml /etc/myapp/config.yml \
    --yaml /etc/myapp/secrets.yml \
    getv \
    --template-file templates/server.xml.tmpl
    > config/server.xml
```
prior to starting tomcat.

### Secret Management
`clconf` can encrypt and decrypt values as well, similar in nature
to `ansible-vault`.  This allows you to commit your _secrets_ alongside
the code that uses them.  For example, you could create a new config
file:
```
db: 
  url: jdbc.mysql:localhost:3306/mydb
```
Then add your secrets:
```
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
```
db:
  password: wcBMA5B5A4w5Zw+rAQgAJ9bR77oJi0P7X5qtnN+soUCszYTy6VGvNutHInE0QCugyXhVeovm+iPaFo/K5D8IO9QJnRL4D9PCiuqVslhsP54b7Qpep/1R/1HEbw9XNMv+uTh9CQDnT1FMer9i+samZ6poTT5uWMJtdTnwa187V5TUGKQdSwoz82CgQ8zQYq0aI15kZp4VziN9eQV1jrphG2+aJdtyIuIouafuEMSnrRz+bb8xAWu3I1INfEP0MuttTYdoY9W3xEU7L4IGvzhw8rnJPNhkK5LKTtvlOCDpKSs1ESReBHYSPNSAAlKBOTHwZ1MHKnypiWVzGACzq+Yh0K+UGtb8dGRiFhwMAn9jfdLgAeRkS/i2wGBjd3suaPzadgW84a0e4L3g3+FKo+Co4k3c3CHgB+OodVAQ2+LoReD54e6X4HbjH52aGIGkSKbg5eLA4qGv4Dnjwf422VOoqubgTOQV3gjv0NTKLF9IXaFPyhtj4joDyk/hwo8A
  url: jdbc.mysql:localhost:3306/mydb
  username: wcBMA5B5A4w5Zw+rAQgAAH1FM4x/FAjmspKbyHJvvaMwmFjGOMOKIle1oe0tpewzaUaEoYZ2trx8nerbWqtIxf4rnB9kNA2YyKs6CLka1q6jnN2U4KI3EjXQaaf6sL5qg/g3Hlak937Wf8+fK1tpghGuFJXTcRjqOgAyV8LfZtQ7MDfgoIy30bihjQz/0TzNi3IZlezqsgvLqoRsgP4b5S9liR/8EaQQ9BepaAgjl3c37QJf/qQK1mkPTOGzlTzZ7dcicpycxRwU8mMlYMq4qN0RR8ZMuiPshYJOdb3OVbNZq08MVzRbuMcPo+SbJsckD+V7EvOn3Km7jefblZsx2fzRPrAG23zZYkAPsUUuE9LgAeTO9rtOh0NQhkYL+9nJzCE+4dpv4K3gCOGkxOBR4o5q737gIuOVjW3r5vC/cuCA4ciT4JDjUV+uW8+IzSfgceKckR304HrjfbEkfn2gljvgAuSCU2yJMaO1aVjs225Rhw7q4pq3xL3hDV4A
```
These values can be decrypted using:
```
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
```
clconf \
    --secret-keyring testdata/test.secring.gpg \
    --yaml C:/Temp/config.yml \
    getv /
    --template-string '{{ cgetv "/db/username" }}:{{ cgetv "/db/password" }}'
```