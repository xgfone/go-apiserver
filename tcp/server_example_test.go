// Copyright 2021 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tcp

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/xgfone/go-apiserver/cert"
	"github.com/xgfone/go-apiserver/internal/test"
)

type logHandler struct {
	name    string
	handler Handler
}

func (h logHandler) OnConnection(c net.Conn) {
	log.Printf("%s: before the new connection from '%s'", h.name, c.RemoteAddr().String())
	h.handler.OnConnection(c)
	log.Printf("%s: after the new connection from '%s'", h.name, c.RemoteAddr().String())
}

func (h logHandler) OnServerExit(err error) {
	log.Printf("%s: before the server exit: %s", h.name, err.Error())
	h.handler.OnServerExit(err)
	log.Printf("%s: after the server exit: %s", h.name, err.Error())
}

func (h logHandler) OnShutdown(c context.Context) {
	log.Printf("%s: before the server shutdown", h.name)
	h.handler.OnShutdown(c)
	log.Printf("%s: after the server shutdown", h.name)
}

func newLogMiddleware(name string) Middleware {
	return func(handler Handler) Handler {
		return logHandler{name: name, handler: handler}
	}
}

func ExampleServer() {
	// Create the certificate.
	cert, err := cert.NewCertificate([]byte(test.Ca), []byte(test.Key), []byte(test.Cert))
	if err != nil {
		log.Fatal(err)
	}

	// Listen the server on the address "127.0.0.1:80"
	ln, err := Listen("127.0.0.1:80")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	// HTTP Router to handle the HTTP request.
	httpHandler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})

	// Initialize the handler and the server.
	httpServerHandler := NewHTTPServerHandler(ln.Addr(), httpHandler)
	/// TODO: set the http server attributions.
	// httpServerHandler.Server.ReadTimeout = time.Second * 10
	// httpServerHandler.Server.WriteTimeout = time.Second * 10
	// httpServerHandler.Server.MaxHeaderBytes = 1024 * 100 // 100KB
	middlewareHandler := NewMiddlewareHandler(httpServerHandler)
	middlewareHandler.Append(newLogMiddleware("mw1"), newLogMiddleware("mw2"))
	server := NewServer(ln, middlewareHandler, cert.TLSConfig)

	// Start the server
	go server.Start()
	go httpServerHandler.Start()

	// Stop the server before the program exits.
	// And httpServerHandler will stop when the tcp server is stopped.
	defer server.Stop()

	// Initialize the HTTP client
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = cert.TLSConfig
	transport.ForceAttemptHTTP2 = false
	httpclient := &http.Client{Transport: transport}

	// Get the request by HTTP
	resp, err := httpclient.Get("http://127.0.0.1:80")
	if resp != nil {
		resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Printf("unexpected the status code '%d'", resp.StatusCode)
		}
	}
	if err != nil {
		log.Println(err)
	}

	// Get the request by HTTPS
	resp, err = httpclient.Get("https://127.0.0.1:80")
	if resp != nil {
		resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Printf("unexpected the status code '%d'", resp.StatusCode)
		}
	}
	if err != nil {
		log.Println(err)
	}

	// Maybe Output:
	// 2021/11/16 17:30:50 mw1: before the new connection from '127.0.0.1:60972'
	// 2021/11/16 17:30:50 mw2: before the new connection from '127.0.0.1:60972'
	// 2021/11/16 17:30:50 mw2: after the new connection from '127.0.0.1:60972'
	// 2021/11/16 17:30:50 mw1: after the new connection from '127.0.0.1:60972'
	// 2021/11/16 17:30:50 mw1: before the new connection from '127.0.0.1:60973'
	// 2021/11/16 17:30:50 mw2: before the new connection from '127.0.0.1:60973'
	// 2021/11/16 17:30:50 mw2: after the new connection from '127.0.0.1:60973'
	// 2021/11/16 17:30:50 mw1: after the new connection from '127.0.0.1:60973'
	// 2021/11/16 17:30:50 mw1: before the server shutdown
	// 2021/11/16 17:30:50 mw2: before the server shutdown
	// 2021/11/16 17:30:50 mw1: before the server exit: accept tcp 127.0.0.1:80: use of closed network connection
	// 2021/11/16 17:30:50 mw2: before the server exit: accept tcp 127.0.0.1:80: use of closed network connection
	// 2021/11/16 17:30:50 mw2: after the server exit: accept tcp 127.0.0.1:80: use of closed network connection
	// 2021/11/16 17:30:50 mw1: after the server exit: accept tcp 127.0.0.1:80: use of closed network connection
	// 2021/11/16 17:30:50 mw2: after the server shutdown
	// 2021/11/16 17:30:50 mw1: after the server shutdown
}
