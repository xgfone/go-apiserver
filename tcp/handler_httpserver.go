// Copyright 2022 xgfone
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
	"fmt"
	"net"
	"net/http"
	"sync/atomic"

	"github.com/xgfone/go-apiserver/http/handler"
	"github.com/xgfone/go-apiserver/log"
)

var _ Handler = &HTTPServerHandler{}

// HTTPServerHandler is a handler to handle the http request.
type HTTPServerHandler struct {
	server   atomic.Value // http.Server
	handler  handler.SwitchHandler
	listener *ForwardConnListener
}

// NewHTTPServerHandler returns a new handler based on the HTTP server.
func NewHTTPServerHandler(localAddr net.Addr, handler http.Handler) *HTTPServerHandler {
	if localAddr == nil {
		panic("HTTPServerHandler: the net listener is nil")
	}
	if handler == nil {
		panic("HTTPServerHandler: the http handler is nil")
	}

	prefix := fmt.Sprintf("HTTPServer(%s): ", localAddr.String())
	logger := log.StdLogger(prefix, log.LvlError)

	h := new(HTTPServerHandler)
	h.handler.Set(handler)
	h.server.Store(&http.Server{Handler: h, ErrorLog: logger})
	h.listener = NewForwardConnListener(localAddr, -1)
	h.listener.OnShutdownFunc = h.onShutdown

	return h
}

// GetHTTPServer returns the inner http server.
func (h *HTTPServerHandler) getHTTPServer() *http.Server {
	return h.server.Load().(*http.Server)
}

func (h *HTTPServerHandler) onShutdown(c context.Context) {
	h.getHTTPServer().Shutdown(c)
}

// ObserveConnState sets the callback function to observe the state of the http
// connection.
func (h *HTTPServerHandler) ObserveConnState(f func(net.Conn, http.ConnState)) {
	h.getHTTPServer().ConnState = f
}

// ServeHTTP implements the interface http.Handler to be used as a http.Handler.
func (h *HTTPServerHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h.Get().ServeHTTP(rw, r)
}

// Get returns the current HTTP handler.
func (h *HTTPServerHandler) Get() http.Handler { return h.handler.Get() }

// Swap stores the new http handler and returns the old.
func (h *HTTPServerHandler) Swap(new http.Handler) (old http.Handler) {
	return h.handler.Swap(new)
}

// Start starts the inner HTTP server.
func (h *HTTPServerHandler) Start() { h.getHTTPServer().Serve(h.listener) }

// OnConnection implements the interface Handler, which will forward the call
// to the inner handler.
func (h *HTTPServerHandler) OnConnection(c net.Conn) { h.listener.OnConnection(c) }

// OnServerExit implements the interface Handler, which will forward the call
// to the inner handler.
func (h *HTTPServerHandler) OnServerExit(err error) { h.listener.OnServerExit(err) }

// OnShutdown implements the interface Handler, which will forward the call
// to the inner handler.
func (h *HTTPServerHandler) OnShutdown(c context.Context) { h.listener.OnShutdown(c) }
