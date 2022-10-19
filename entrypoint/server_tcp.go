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
	"github.com/xgfone/go-apiserver/tls/tlscert"
)

func init() {
	RegisterServerBuilder("tcp", func(addr string, h interface{}) (Server, error) {
		ln, err := tcp.Listen(addr)
		if err != nil {
			return nil, err
		}
		return NewTCPServer(ln, h.(tcp.Handler)), nil
	})
}

var _ Server = TCPServer{}

// TCPServer represents a tcp entrypoint server.
type TCPServer struct {
	Middlewares *middleware.Manager
	CertManager *tlscert.Manager
	*tcp.Server
}

// NewTCPServer returns a new TCP entrypoint Server.
func NewTCPServer(ln net.Listener, handler tcp.Handler) (server TCPServer) {
	if ln == nil {
		panic("the tcp listener is nil")
	} else if handler == nil {
		panic("the tcp handler is nil")
	}

	server.CertManager = tlscert.NewManager()
	server.Middlewares = middleware.NewManager(handler)
	server.Server = tcp.NewServer(ln, server.Middlewares)
	server.SetTLSForce(true)
	return
}

// Protocol returns the protocol of the http server, which is a fixed "tcp".
func (s TCPServer) Protocol() string { return "tcp" }

// SetTLSConfig sets the tls configuration, which is thread-safe.
func (s TCPServer) SetTLSConfig(c *tls.Config) {
	if c != nil && c.GetCertificate == nil && len(c.Certificates) == 0 {
		c.GetCertificate = s.CertManager.GetTLSCertificate
	}

	_, forceTLS := s.Server.GetTLSConfig()
	s.Server.SetTLSConfig(c, forceTLS)
}

// SetTLSForce sets whether or not to force the client to use TLS.
func (s TCPServer) SetTLSForce(forceTLS bool) {
	config, _ := s.Server.GetTLSConfig()
	s.Server.SetTLSConfig(config, forceTLS)
}

// AddCertificate implements the interface tlscert.CertUpdater.
func (s TCPServer) AddCertificate(name string, certificate tlscert.Certificate) {
	s.CertManager.AddCertificate(name, certificate)
}

// DelCertificate implements the interface tlscert.CertUpdater.
func (s TCPServer) DelCertificate(name string) {
	s.CertManager.DelCertificate(name)
}
