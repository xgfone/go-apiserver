# go-apiserver [![Build Status](https://github.com/xgfone/go-apiserver/actions/workflows/go.yml/badge.svg)](https://github.com/xgfone/go-apiserver/actions/workflows/go.yml) [![GoDoc](https://pkg.go.dev/badge/github.com/xgfone/go-apiserver)](https://pkg.go.dev/github.com/xgfone/go-apiserver) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/go-apiserver/master/LICENSE)

The library to build an API server, such as `API Gateway`, based on `Go1.13+`.


## Features
- CURD are thread-safe and operated during the program is running.
- For the data plane, such as routing the request, it is atomic and not affected by CURD.
- Each package is standalone and used independently. Moreover, they are also combined to work.


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
	"github.com/xgfone/go-apiserver/http/middlewares"
	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/http/router"
	"github.com/xgfone/go-apiserver/http/router/routes/ruler"
)

func httpHandler(route string) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		c := reqresp.GetContext(r)
		fmt.Fprintf(rw, "%s: %s", route, c.Data["id"])
	}
}

// StartHTTPServer is a simple convenient function to start a http server.
func StartHTTPServer(addr string, handler http.Handler) (err error) {
	ep := entrypoint.NewEntryPoint("test", addr, handler)
	if err = ep.Init(); err == nil {
		ep.Start()
	}
	return
}

func main() {
	routeManager := ruler.NewRouteManager()

	// Route 1: Build the route by the matcher rule string
	routeManager.
		Rule("Method(`GET`) && Path(`/path1/{id}`)"). // Build the matcher
		Handler(httpHandler("route1"))                // Set the handler

	// Route 2: Build the route by assembling the matcher.
	routeManager.
		Path("/path2/{id}").Method("GET"). // Assemble the matchers
		Handler(httpHandler("route2"))     // Set the handler

	routeManager.
		Path("/path3/{id}").Method("GET"). // Assemble the matchers
		HandlerFunc(httpHandler("route3")) // Set the handler function

	router := router.NewRouter(routeManager)
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
		fmt.Fprintf(rw, "%s: %s", route, c.Data["id"])
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
	"crypto/tls"
	"flag"
	"fmt"
	"time"

	"github.com/xgfone/go-apiserver/entrypoint"
	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/http/router/routes/ruler"
	"github.com/xgfone/go-apiserver/tls/tlscert/provider"
)

var (
	addr     = flag.String("addr", "127.0.0.1:80", "The address to listen to.")
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
	ep := entrypoint.NewEntryPoint("entrypoint_name", *addr, router)
	if err := ep.Init(); err != nil {
		fmt.Println(err)
		return
	}

	// Initialize the certificate.
	if *keyFile != "" && *certFile != "" {
		// Use the file provider to monitor the change of the certificate files.
		// When the certificate files have changed, it will be reloaded.
		tlsFileProvider := provider.NewFileProvider("tlsfiles", time.Second*10)
		tlsFileProvider.AddCertFile("tlsfile", *keyFile, *certFile)

		// Use the provider manager to manage all the certificate providers,
		// and update the certificate into the entrypoint when it has changed.
		tlsProviderManager := provider.NewManager(ep)
		tlsProviderManager.AddProvider(tlsFileProvider)
		tlsProviderManager.Start(context.Background())

		// Update the TLS configuration of the entrypoint.
		ep.SetTLSConfig(new(tls.Config))
		ep.SetTLSForce(*forceTLS)
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
	// $ curl https://localhost:80 --cacert ca.pem
	// curl: (35) error:1408F10B:SSL routines:ssl3_get_record:wrong version number
	//
	//
	// TEST 2:
	// Open a terminal and run the program:
	// $ go run main.go -keyfile key.pem -certfile cert.pem
	//
	// Open another terminal and run the http client:
	// $ curl http://localhost:80
	// OK
	// $ curl https://localhost:80 --cacert ca.pem
	// OK
	//
	//
	// TEST 3:
	// Open a terminal and run the program:
	// $ go run main.go -keyfile key.pem -certfile cert.pem --forcetls
	//
	// Open another terminal and run the http client:
	// $ curl http://localhost:80
	// curl: (56) Recv failure: Connection was reset
	// $ curl https://localhost:80 --cacert ca.pem
	// OK
}
```

### Mini API Gateway
```go
package main

import (
	"flag"

	"github.com/xgfone/go-apiserver/entrypoint"
	"github.com/xgfone/go-apiserver/http/middlewares"
	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/http/router"
	"github.com/xgfone/go-apiserver/http/router/routes/ruler"
	"github.com/xgfone/go-apiserver/http/upstream"
	"github.com/xgfone/go-apiserver/http/upstream/balancer"
	"github.com/xgfone/go-apiserver/http/upstream/loadbalancer"
	"github.com/xgfone/go-apiserver/log"
)

var (
	listenAddr = flag.String("listen-addr", ":80", "The address that api gateway listens on.")
	manageCIDR = flag.String("manage-cidr", "127.0.0.0/8", "The CIDR of the management network.")
)

func main() {
	flag.Parse()

	routeManager := ruler.NewRouteManager()
	initAdminManageAPI(routeManager)

	router := router.NewRouter(routeManager)
	router.Middlewares.Use(middlewares.DefaultMiddlewares...)

	ep := entrypoint.NewEntryPoint("api-gateway", *listenAddr, router)
	if err := ep.Init(); err != nil {
		log.Error("fail to initialize the api-gateway entrypoint", "addr", *listenAddr)
		return
	}

	ep.Start()
}

func initAdminManageAPI(router *ruler.RouteManager) {
	router.
		Path("/admin/route").
		Method("POST").
		ClientIP(*manageCIDR). // Only allow the specific clients to add the route.
		ContextHandler(func(ctx *reqresp.Context) {
			var req struct {
				Rule     string `json:"rule"`
				Upstream struct {
					ForwardPolicy string       `json:"forwardPolicy"`
					ForwardURL    upstream.URL `json:"forwardUrl"`

					Servers []struct {
						IP     string `json:"ip"`
						Port   uint16 `json:"port"`
						Weight int    `json:"weight"`
					} `json:"servers" yaml:"servers"`
				} `json:"upstream"`
			}

			if err := ctx.BindBody(&req); err != nil {
				ctx.Text(400, "invalid request route paramenter: %s", err.Error())
				return
			}

			if req.Rule == "" {
				ctx.Text(400, "the route rule is empty")
				return
			}
			if req.Upstream.ForwardPolicy == "" {
				req.Upstream.ForwardPolicy = "weight_random"
			}

			// Build the upstream servers.
			servers := make(upstream.Servers, len(req.Upstream.Servers))
			for i, server := range req.Upstream.Servers {
				config := upstream.ServerConfig{URL: req.Upstream.ForwardURL}
				config.StaticWeight = server.Weight
				config.Port = server.Port
				config.IP = server.IP

				wserver, err := upstream.NewServer(config)
				if err != nil {
					ctx.Text(400, "fail to build the upstream server: %s", err.Error())
					return
				}

				servers[i] = wserver
			}

			// Build the loadbalancer forwarder.
			balancer, _ := balancer.Build(req.Upstream.ForwardPolicy, nil)
			lb := loadbalancer.NewLoadBalancer(req.Rule, balancer)
			lb.UpsertServers(servers...)

			// Build the route and forward the request to loadbalancer.
			err := router.Rule(req.Rule).HandlerFunc(lb.ServeHTTP)
			if err != nil {
				ctx.Text(400, "invalid route rule '%s': %s", req.Rule, err.Error())
			}
		})
}
```

```shell
# Run the mini API-Gateway on the host 192.168.1.10
$ nohup go run main.go &

# Add the route
$ curl -XPOST http://127.0.0.1/admin/route -H 'Content-Type: application/json' -d '
{
    "rule": "Method(`GET`) && Path(`/path`)",
    "upstream": {
        "forwardPolicy": "weight_round_robin",
        "forwardUrl" : {"path": "/backend/path"},
        "servers": [
            {"ip": "192.168.1.11", "port": 80, "weight": 1}, // 33.3% requests
            {"ip": "192.168.1.12", "port": 80, "weight": 2}  // 66.7% requests
        ]
    }
}'

# Access the backend servers by the mini API-Gateway:
# 1/3 requests -> 192.168.1.11
# 2/3 requests -> 192.168.1.12
$ curl http://192.168.1.10/path
192.168.1.11/backend/path

$ curl http://192.168.1.10/path
192.168.1.12/backend/path

$ curl http://192.168.1.10/path
192.168.1.12/backend/path

$ curl http://192.168.1.10/path
192.168.1.11/backend/path

$ curl http://192.168.1.10/path
192.168.1.12/backend/path

$ curl http://192.168.1.10/path
192.168.1.12/backend/path
```
