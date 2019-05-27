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

## post package DB scripts

Since v0.3.0 of the github.com/jecolon/post package, SQLite3 is used for persistence of posts in the file posts.db.
The file will be created at the project root of the server (outside the webroot for security) and the following scripts are
provided for testing and maintenace.

* dbreset.sh: Drops the post table, leaving the DB empty.
* dbfill.sh: Creates post table if necessary and inserts 10 posts for testing.
