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

package entrypoint

import (
	"crypto/tls"
	"net"

	"github.com/xgfone/go-apiserver/tcp"
	"github.com/xgfone/go-apiserver/tcp/middleware"
)

var _ Server = TCPServer{}

// TCPServer represents a tcp entrypoint server.
type TCPServer struct {
	Middlewares *middleware.Manager
	*tcp.Server
}

// NewTCPServer returns a new TCP entrypoint Server.
func NewTCPServer(ln net.Listener, handler tcp.Handler) (server TCPServer) {
	if ln == nil {
		panic("the tcp listener is nil")
	}
	if handler == nil {
		panic("the tcp handler is nil")
	}

	server.Middlewares = middleware.NewManager(handler)
	server.Server = tcp.NewServer(ln, server.Middlewares, nil)
	return
}

// Protocal returns the protocal of the http server, which is a fixed "tcp".
func (s TCPServer) Protocal() string { return "tcp" }

// SetTLSConfig sets the TLS configuration.
func (s TCPServer) SetTLSConfig(config *tls.Config, forceTLS bool) {
	s.Server.TLSConfig = config
	s.Server.ForceTLS = forceTLS
}
