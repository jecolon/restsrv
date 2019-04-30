# restsrv
A sample HTTP2 REST API Server.

## Quick Start
You need the latest version of Go https://golang.org

```bash
$ git clone https://github.com/jecolon/restsrv
$ cd restsrv
$ mkdir tls
$ cd tls
$ go run /usr/local/go/src/crypto/tls/generate_cert.go --host localhost
$ cd ..
$ go build
$ ./restsrv
```
Navigate to https://localhost:8443 for file server homepage and https://localhost:8443/api/v1/posts for REST API using 
github.com/jecolon/post package.
