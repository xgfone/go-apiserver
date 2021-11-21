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

// Package tcp implements a TCP server and some handlers.
package tcp

import (
	"context"
	"crypto/tls"
	"errors"
	"net"

	"github.com/xgfone/go-apiserver/log"
)

// TLSConn is used to support to check whether the connection is based on TLS.
type TLSConn interface {
	// TLSConn returns the TLS connection if it is based on TLS. Or, reutrn nil.
	TLSConn() *tls.Conn

	net.Conn
}

// Handler is used to handle the TCP connection.
type Handler interface {
	// OnConnection is called when a new connection comes.
	//
	// For the TCP connection, you can assert the connection to TLSConn
	// to get the TLS connection.
	//
	// Notice: it is the responsibility of the handler to close the connection.
	OnConnection(net.Conn)

	// OnServerExit is called when the server exits unexpectedly.
	OnServerExit(err error)

	// OnShutdown is called when the server is stopped.
	OnShutdown(context.Context)
}

// Server implements a server based on the stream.
type Server struct {
	Handler   Handler
	Listener  net.Listener
	TLSConfig *tls.Config
}

// NewServer returns a new Server.
func NewServer(ln net.Listener, handler Handler, config *tls.Config) *Server {
	return &Server{Listener: ln, Handler: handler, TLSConfig: config}
}

// Stop stops the server and waits until all the connections are closed.
func (s *Server) Stop() { s.Shutdown(context.Background()) }

// Shutdown shuts down the server gracefully.
func (s *Server) Shutdown(ctx context.Context) {
	s.Listener.Close()
	s.Handler.OnShutdown(ctx)
}

// Start starts the TCP server.
func (s *Server) Start() {
	addr := log.F("addr", s.Listener.Addr().String())

	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			log.Error("failed to accept the new connection", addr, log.E(err))

			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Temporary() {
				continue
			}

			s.Handler.OnServerExit(err)
			return
		}

		if s.TLSConfig != nil {
			conn = &tlsConn{Conn: conn, config: s.TLSConfig, first: true}
		}

		s.Handler.OnConnection(conn)
	}
}

type tlsConn struct {
	config *tls.Config
	istls  bool
	first  bool
	err    error

	net.Conn
}

var _ TLSConn = &tlsConn{}

func (c *tlsConn) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return
	}

	c.ensureTLSConn()
	if c.err != nil {
		err, c.err = c.err, nil
		return
	}

	n, err = c.Conn.Read(p)
	return
}

func (c *tlsConn) TLSConn() (conn *tls.Conn) {
	c.ensureTLSConn()
	if c.istls {
		conn, _ = c.Conn.(*tls.Conn)
	}
	return
}

func (c *tlsConn) ensureTLSConn() {
	// For first read, We will detect whether the connection is based on TLS.
	if c.first {
		c.first = false

		var bs [1]byte
		if _, c.err = c.Conn.Read(bs[:]); c.err != nil {
			log.Error("fail to read the first byte from the tcp conneciton",
				log.F("remoteaddr", c.RemoteAddr().String()), log.E(c.err))
			c.Close()
			return
		}

		c.Conn = &peekedConn{Conn: c.Conn, Peeked: int16(bs[0])}

		// Detect whether the connection is based on TLS. If true, use tls.Server
		// to wrap the original connection.
		//
		// No valid TLS record has a type of 0x80, however SSLv2 handshakes
		// start with a uint16 length where the MSB is set and the first record
		// is always < 256 bytes long. Therefore typ == 0x80 strongly suggests
		// an SSLv2 client.
		const recordTypeSSLv2 = 0x80
		const recordTypeHandshake = 0x16
		if bs[0] == recordTypeHandshake || bs[0] == recordTypeSSLv2 { // For TLS
			c.Conn = tls.Server(c.Conn, c.config)
			c.istls = true
		}
	}
}

type peekedConn struct {
	Peeked int16
	net.Conn
}

func (c *peekedConn) Read(p []byte) (n int, err error) {
	if c.Peeked > -1 {
		p[0] = byte(c.Peeked)
		c.Peeked = -1
		if len(p) == 1 {
			return 1, nil
		}

		p = p[1:]
		n = 1
	}

	m, err := c.Conn.Read(p)
	n += m
	return
}
