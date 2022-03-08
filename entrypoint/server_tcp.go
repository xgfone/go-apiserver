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

	"github.com/xgfone/go-apiserver/middleware"
	"github.com/xgfone/go-apiserver/tcp"
	"github.com/xgfone/go-apiserver/tlscert"
)

var _ Server = TCPServer{}

// TCPServer represents a tcp entrypoint server.
type TCPServer struct {
	ServerName  string
	Middlewares *middleware.Manager
	*tlscert.CertManager
	*tcp.Server
}

// NewTCPServer returns a new TCP entrypoint Server.
func NewTCPServer(serverName string, ln net.Listener, handler tcp.Handler) (server TCPServer) {
	if ln == nil {
		panic("the tcp listener is nil")
	}
	if handler == nil {
		panic("the tcp handler is nil")
	}

	server.ServerName = serverName
	server.CertManager = tlscert.NewCertManager(serverName)
	server.Middlewares = middleware.NewManager(handler)
	server.Server = tcp.NewServer(ln, server.Middlewares)
	return
}

// Name returns the name of the tcp server.
func (s TCPServer) Name() string { return s.ServerName }

// Protocal returns the protocal of the http server, which is a fixed "tcp".
func (s TCPServer) Protocal() string { return "tcp" }

// SetTLSConfig sets the tls configuration, which is thread-safe.
func (s TCPServer) SetTLSConfig(tlsConfig *tls.Config, forceTLS bool) {
	if tlsConfig.GetCertificate == nil && len(tlsConfig.Certificates) == 0 {
		tlsConfig.GetCertificate = s.CertManager.GetTLSCertificate
	}
	s.Server.SetTLSConfig(tlsConfig, forceTLS)
}
