# go-apiserver [![Build Status](https://github.com/xgfone/go-apiserver/actions/workflows/go.yml/badge.svg)](https://github.com/xgfone/go-apiserver/actions/workflows/go.yml) [![GoDoc](https://pkg.go.dev/badge/github.com/xgfone/go-apiserver)](https://pkg.go.dev/github.com/xgfone/go-apiserver) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/go-apiserver/master/LICENSE)

The library to build an API server, such as `API Gateway`, based on `Go1.13+`.


## Install
```shell
$ go get -u github.com/xgfone/go-apiserver
```


## Example

### Simple HTTP Router
```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/xgfone/go-apiserver/entrypoint"
	"github.com/xgfone/go-apiserver/http/middleware/middlewares"
	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/http/router"
	"github.com/xgfone/go-apiserver/http/router/routes/ruler"
)

func httpHandler(route string) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		c := reqresp.GetContext(r)
		fmt.Fprintf(rw, "%s: %s", route, c.Datas["id"])
	}
}

// StartHTTPServer is a simple convenient function to start a http server.
func StartHTTPServer(addr string, handler http.Handler) error {
	ep, err := entrypoint.NewHTTPEntryPoint("test", addr, handler)
	if err != nil {
		return err
	}

	ep.Start()
	return nil
}

func main() {
	routers := ruler.NewRouteManager()

	// Route 1: Build the route by the matcher rule string
	routers.
		Rule("Method(`GET`) && Path(`/path1/{id}`)"). // Build the matcher
		Handler(httpHandler("route1"))                // Set the handler

	// Route 2: Build the route by assembling the matcher.
	routers.
		Path("/path2/{id}").Method("GET"). // Assemble the matchers
		Handler(httpHandler("route2"))     // Set the handler

	routers.
		Path("/path3/{id}").Method("GET"). // Assemble the matchers
		HandlerFunc(httpHandler("route3")) // Set the handler function

	router := router.NewRouter(routers)
	router.Middlewares.Use(middlewares.Logger(0), middlewares.Recover(1))

	// err := http.ListenAndServe("127.0.0.1:80", router)
	err := StartHTTPServer("127.0.0.1:80", router)
	log.Printf("http server shutdown: %s", err)

	// Open a terminal and run the program:
	// $ go run main.go
	//
	// Open another terminal and run the http client:
	// $ curl http://127.0.0.1/path1/123
	// route1: 123
	// $ curl http://127.0.0.1/path2/123
	// route2: 123
	// $ curl http://127.0.0.1/path3/123
	// route3: 123
}
```

### Virtual Host
```go
package main

import (
	"fmt"
	"net/http"

	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/http/router/routes/ruler"
	"github.com/xgfone/go-apiserver/http/vhost"
)

func httpHandler(route string) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		c := reqresp.GetContext(r)
		fmt.Fprintf(rw, "%s: %s", route, c.Datas["id"])
	}
}

func main() {
	vhosts := vhost.NewManager()

	// Set the default
	defaultVhost := ruler.NewRouteManager()
	defaultVhost.Path("/path/{id}").GET(httpHandler("default"))
	vhosts.SetDefaultVHost(defaultVhost)
	// vhosts.SetDefaultVHost(httpHandler("default")) // Use the http handler as vhost.

	// VHost 1: www.example1.com
	vhost1 := ruler.NewRouteManager()
	vhost1.Path("/path/{id}").GET(httpHandler("route1"))
	vhosts.AddVHost("www.example1.com", vhost1)

	// VHost 2: www.example2.com
	vhost2 := ruler.NewRouteManager()
	vhost2.Path("/path/{id}").GET(httpHandler("route2"))
	vhosts.AddVHost("www.example2.com", vhost2)

	// VHost 3: *.example1.com
	vhost3 := ruler.NewRouteManager()
	vhost3.Path("/path/{id}").GET(httpHandler("route3"))
	vhosts.AddVHost("*.example1.com", vhost3)

	// VHost 4: www.example2.com
	vhost4 := ruler.NewRouteManager()
	vhost4.Path("/path/{id}").GET(httpHandler("route4"))
	vhosts.AddVHost("*.example2.com", vhost4)

	http.ListenAndServe("127.0.0.1:80", vhosts)

	// Open a terminal and run the program:
	// $ go run main.go
	//
	// Open another terminal and run the http client:
	// $ curl http://127.0.0.1/path/123 -H 'Host: www.example1.com'
	// route1: 123
	// $ curl http://127.0.0.1/path/123 -H 'Host: www.example2.com'
	// route2: 123
	// $ curl http://127.0.0.1/path/123 -H 'Host: abc.example1.com'
	// route3: 123
	// $ curl http://127.0.0.1/path/123 -H 'Host: abc.example2.com'
	// route4: 123
	// $ curl http://127.0.0.1/path/123 -H 'Host: www.example.com'
	// default: 123
}
```

### TLS Certificate
```go
package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/xgfone/go-apiserver/entrypoint"
	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/http/router/routes/ruler"
	"github.com/xgfone/go-apiserver/tlscert"
)

var (
	addr     = flag.String("addr", ":80", "The address to listen to.")
	caFile   = flag.String("cafile", "", "The CA file.")
	keyFile  = flag.String("keyfile", "", "The Key file.")
	certFile = flag.String("certfile", "", "The Certificate file.")
	forceTLS = flag.Bool("forcetls", false, "If true, use TLS forcibly when setting the certificate")
)

func main() {
	// Parse the CLI arguments.
	flag.Parse()

	// New the http router.
	router := ruler.NewRouteManager()
	router.Path("/").ContextHandler(func(c *reqresp.Context) { c.Text(200, "OK") })

	// New an entrypoint based on HTTP.
	ep, err := entrypoint.NewHTTPEntryPoint("entrypoint_name", *addr, router)
	if err != nil {
		fmt.Println(err)
		return
	}

	if *keyFile != "" && *certFile != "" {
		// Use the certificate manager to manage all the certificates.
		certManager := tlscert.NewCertManager("tlsgroup")

		// Use the file provider to monitor the change of the certificate files.
		// When the certificate files have changed, it will be reloaded.
		tlsFileProvider := tlscert.NewFileProvider("tlsfiles", time.Second*10)
		tlsFileProvider.AddCertFile("tlsfile", *caFile, *keyFile, *certFile)

		// Use the provider manager to manage all the certificate providers.
		tlsProviderManager := tlscert.NewProviderManager(certManager)
		tlsProviderManager.AddProvider(tlsFileProvider)
		tlsProviderManager.Start(context.Background())

		// Set the TLS configuration of the entrypoint.
		ep.TLSConfig = certManager.TLSConfig()
		ep.ForceTLS = *forceTLS
	}

	// Start the HTTP server.
	ep.Start()

	// TEST 1:
	// Open a terminal and run the program:
	// $ go run main.go
	//
	// Open another terminal and run the http client:
	// $ curl http://localhost:80
	// OK
	// $ curl https://localhost:80 --cacert ca.pem --key key.pem --cert cert.pem
	// curl: (35) error:1408F10B:SSL routines:ssl3_get_record:wrong version number
	//
	//
	// TEST 2:
	// Open a terminal and run the program:
	// $ go run main.go -cafile ca.pem -keyfile key.pem -certfile cert.pem
	//
	// Open another terminal and run the http client:
	// $ curl http://localhost:80
	// OK
	// $ curl https://localhost:80 --cacert ca.pem --key key.pem --cert cert.pem
	// OK
	//
	//
	// TEST 3:
	// Open a terminal and run the program:
	// $ go run main.go -cafile ca.pem -keyfile key.pem -certfile cert.pem --forcetls
	//
	// Open another terminal and run the http client:
	// $ curl http://localhost:80
	// curl: (56) Recv failure: Connection was reset
	// $ curl https://localhost:80 --cacert ca.pem --key key.pem --cert cert.pem
	// OK
}
```
