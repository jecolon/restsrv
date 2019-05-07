# restsrv
A sample HTTP2 REST API Server.

## Quick Start
You need the latest version of Go https://golang.org

```bash
$ git clone https://github.com/jecolon/restsrv
$ cd restsrv
$ mkdir -p tls/{dev, prod}
$ cd tls/dev
$ go run /usr/local/go/src/crypto/tls/generate_cert.go --host localhost
$ cd ../..
$ go build
$ ./restsrv -d
```

Navigate to https://localhost:8443 for file server homepage and 
https://localhost:8443/api/v1/posts for REST API using github.com/jecolon/post package.

## Options

* -d : For local dev mode using TLS cert.pem and key.pem from tls/dev directory.
Without this option, server runs in production mode looking for these files in 
tls/prod instead. (default: false)
* -w [webroot] : Specifies the webroot for the file server. (default: "webroot")
* -p [port] : Specifies the IP address and port to listen on. (default: ":8443")
* -h or -? : Show usage information.

