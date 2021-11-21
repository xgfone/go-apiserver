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

// Package entrypoint implements the http entrypoint and the manager.
package entrypoint

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	routerhttp "github.com/xgfone/go-apiserver/router/http"
	"github.com/xgfone/go-apiserver/server/tcp"
)

type server interface {
	Shutdown(context.Context)
	Start()
	Stop()
}

var _ server = &EntryPoint{}

// var nothing = http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})

// EntryPoint represents an entrypoint of the services.
type EntryPoint struct {
	// The unique name of the entrypoint.
	Name string

	// The address that the entrypoint listens on.
	//
	// Format: [(http|tcp|udp)://][host]:port
	//
	// If missing the protocol, it is "http" by default.
	Addr   string
	Config *tls.Config

	server      server
	protocol    string
	mwHandler   *tcp.MiddlewareHandler
	httpHandler *tcp.HTTPServerHandler
}

// NewEntryPoint returns a new entrypoint.
func NewEntryPoint(name, addr string) (*EntryPoint, error) {
	ep := EntryPoint{Name: name, Addr: addr, protocol: "http"}

	if index := strings.Index(addr, "://"); index > -1 {
		ep.protocol = addr[:index]
		addr = addr[index+3:]
	}

	switch ep.protocol {
	case "", "http":
		ln, err := tcp.Listen(addr)
		if err != nil {
			return nil, err
		}

		ep.httpHandler = tcp.NewHTTPServerHandler(ln.Addr(), routerhttp.Handler404)
		ep.mwHandler = tcp.NewMiddlewareHandler(ep.httpHandler)
		ep.server = tcp.NewServer(ln, ep.mwHandler, nil)

	// case "tcp":
	// case "udp":
	default:
		return nil, fmt.Errorf("unknown entrypoint protocol '%s'", ep.protocol)
	}

	return &ep, nil
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
func (ep *EntryPoint) AppendTCPHandlerMiddlewares(mws ...tcp.Middleware) {
	ep.mwHandler.Append(mws...)
}

// Start starts the entrypoint.
func (ep *EntryPoint) Start() {
	switch server := ep.server.(type) {
	case *tcp.Server:
		server.TLSConfig = ep.Config

	// case *udp.Server:
	default:
		panic("unknown the entrypoint server type")
	}

	go ep.httpHandler.Start()
	ep.server.Start()
}

// Stop stops the entrypoint and waits until all the connections are closed.
func (ep EntryPoint) Stop() { ep.server.Stop() }

// Shutdown shuts down the entrypoint gracefully.
func (ep EntryPoint) Shutdown(c context.Context) { ep.server.Shutdown(c) }
