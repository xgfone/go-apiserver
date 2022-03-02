// Copyright 2021~2022 xgfone
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

// Package entrypoint implements the http entrypoint and the manager.
package entrypoint

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	"github.com/xgfone/go-apiserver/http/router"
	"github.com/xgfone/go-apiserver/http/router/routes/ruler"
	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/tcp"
	"github.com/xgfone/go-apiserver/tcp/middleware"
)

type server interface {
	OnShutdown(...func())
	Shutdown(context.Context)
	Start()
	Stop()
}

var _ server = &EntryPoint{}

// EntryPoint represents an entrypoint of the services.
type EntryPoint struct {
	// The unique name of the entrypoint.
	Name string

	// The address that the entrypoint listens on.
	//
	// Format: [(http|tcp|udp)://][host]:port
	//
	// If missing the protocol, it is "http" by default.
	Addr string

	// TLSConfig is used to configure the TLS.
	TLSConfig *tls.Config
	ForceTLS  bool

	server      server
	protocol    string
	mwHandler   *middleware.Manager
	httpHandler *tcp.HTTPServerHandler
}

// NewHTTPEntryPoint returns a new http entrypoint.
func NewHTTPEntryPoint(name, addr string, handler http.Handler) (*EntryPoint, error) {
	ln, err := tcp.Listen(addr)
	if err != nil {
		return nil, err
	}

	if handler == nil {
		handler = router.NewRouter(ruler.NewRouteManager())
	}

	ep := &EntryPoint{Name: name, Addr: addr, protocol: "http"}
	ep.httpHandler = tcp.NewHTTPServerHandler(ln.Addr(), handler)
	ep.mwHandler = middleware.NewManager(ep.httpHandler)
	ep.server = tcp.NewServer(ln, ep.mwHandler, nil)
	return ep, nil
}

// NewEntryPoint returns a new entrypoint.
func NewEntryPoint(name, addr string) (*EntryPoint, error) {
	var protocol string
	if index := strings.Index(addr, "://"); index > -1 {
		protocol = addr[:index]
		addr = addr[index+3:]
	}

	switch protocol {
	case "", "http":
		return NewHTTPEntryPoint(name, addr, nil)

	// case "tcp":
	// case "udp":
	default:
		return nil, fmt.Errorf("unknown entrypoint protocol '%s'", protocol)
	}
}

// Protocol returns the protocol of the entrypoint, such as "http", "tcp" or "udp".
func (ep *EntryPoint) Protocol() string { return ep.protocol }

// GetHTTPHandler returns the http handler.
func (ep *EntryPoint) GetHTTPHandler() http.Handler {
	return ep.httpHandler.Get()
}

// SwitchHTTPHandler swaps out the old http handler with the new.
func (ep *EntryPoint) SwitchHTTPHandler(new http.Handler) (old http.Handler) {
	return ep.httpHandler.Swap(new)
}

// AppendTCPHandlerMiddlewares appends the tcp handler middlewares.
func (ep *EntryPoint) AppendTCPHandlerMiddlewares(mws ...middleware.Middleware) {
	ep.mwHandler.Use(mws...)
}

// OnShutdown registers the callback functions, which are called
// when the entrypoint is shut down.
func (ep *EntryPoint) OnShutdown(callbacks ...func()) {
	ep.server.OnShutdown(callbacks...)
}

// Start starts the entrypoint.
func (ep *EntryPoint) Start() {
	switch server := ep.server.(type) {
	case *tcp.Server:
		server.TLSConfig = ep.TLSConfig
		server.ForceTLS = ep.ForceTLS

	// case *udp.Server:
	default:
		panic("unknown the entrypoint server type")
	}

	log.Info(fmt.Sprintf("start the %s server", ep.protocol),
		"enabletls", ep.TLSConfig != nil, "forcetls", ep.ForceTLS,
		"name", ep.Name, "listenaddr", ep.Addr)

	go ep.httpHandler.Start()
	ep.server.Start()

	log.Info(fmt.Sprintf("stop the %s server", ep.protocol),
		"enabletls", ep.TLSConfig != nil, "forcetls", ep.ForceTLS,
		"name", ep.Name, "listenaddr", ep.Addr)
}

// Stop stops the entrypoint and waits until all the connections are closed.
func (ep *EntryPoint) Stop() { ep.server.Stop() }

// Shutdown shuts down the entrypoint gracefully.
func (ep *EntryPoint) Shutdown(c context.Context) { ep.server.Shutdown(c) }
