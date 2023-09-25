# go-apiserver [![Build Status](https://github.com/xgfone/go-apiserver/actions/workflows/go.yml/badge.svg)](https://github.com/xgfone/go-apiserver/actions/workflows/go.yml) [![GoDoc](https://pkg.go.dev/badge/github.com/xgfone/go-apiserver)](https://pkg.go.dev/github.com/xgfone/go-apiserver) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/go-apiserver/master/LICENSE)


The library is used to build an API server, requiring `Go1.21+`.


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
	"net/http"

	"github.com/xgfone/go-apiserver/http/middleware"
	"github.com/xgfone/go-apiserver/http/middleware/context"
	"github.com/xgfone/go-apiserver/http/middleware/logger"
	"github.com/xgfone/go-apiserver/http/middleware/recover"
	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/http/router"
	"github.com/xgfone/go-apiserver/http/router/ruler"
	"github.com/xgfone/go-apiserver/http/server"
)

func httpHandler(route string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := reqresp.GetContext(r.Context())
		fmt.Fprintf(w, "%s: %s", route, c.Data["id"])
	}
}

func main() {
	// Build the routes.
	routeManager := ruler.NewRouter()
	routeManager.Path("/path1/{id}").Method("GET").Handler(httpHandler("route1"))
	routeManager.Path("/path2/{id}").GET(httpHandler("route2"))

	router := router.NewRouter(routeManager)
	router.Use(middleware.MiddlewareFunc(context.Context)) // Use a function as middleware
	router.Use(
		middleware.New("logger", 1, logger.Logger),    // Use the named priority middleware
		middleware.New("recover", 2, recover.Recover), // Use the named priority middleware
	)

	// http.ListenAndServe("127.0.0.1:80", router)
	server.Start("127.0.0.1:80", router)

	// Open a terminal and run the program:
	// $ go run main.go
	//
	// Open another terminal and run the http client:
	// $ curl http://127.0.0.1/path1/123
	// route1: 123
	// $ curl http://127.0.0.1/path2/123
	// route2: 123
}
```
