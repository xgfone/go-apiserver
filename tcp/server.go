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
	"crypto/tls"
	"errors"
	"net"
	"sync/atomic"

	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/nets"
)

// TLSConn is used to support to check whether the connection is based on TLS.
type TLSConn interface {
	// TLSConn returns the TLS connection if it is based on TLS. Or, reutrn nil.
	TLSConn() *tls.Conn

	net.Conn
}

// Server implements a server based on the stream.
type Server struct {
	Handler  Handler
	Listener net.Listener

	// If TLSConfig is set and ForeceTLS is true, the client must use TLS.
	// If TLSConfig is set and ForeceTLS is false, the client maybe use TLS or not-TLS.
	// If TLSConfig is not set, ForceTLS is ignored and the client must use not-TLS.
	TLSConfig *tls.Config // Default: nil
	ForceTLS  bool        // Default: false

	stopped int32
	stops   []func()
}

// NewServer returns a new Server.
func NewServer(ln net.Listener, handler Handler, config *tls.Config) *Server {
	return &Server{Listener: ln, Handler: handler, TLSConfig: config}
}

func (s *Server) shutdown(ctx context.Context) {
	s.Listener.Close()
	s.Handler.OnShutdown(ctx)
	for _len := len(s.stops) - 1; _len >= 0; _len-- {
		s.stops[_len]()
	}
}

// Stop stops the server and waits until all the connections are closed.
func (s *Server) Stop() { s.Shutdown(context.Background()) }

// Shutdown shuts down the server gracefully.
func (s *Server) Shutdown(ctx context.Context) {
	if atomic.CompareAndSwapInt32(&s.stopped, 0, 1) {
		s.shutdown(ctx)
	}
}

// OnShutdown registers the callback functions, which are called
// when the server is shut down.
func (s *Server) OnShutdown(callbacks ...func()) {
	s.stops = append(s.stops, callbacks...)
}

// Start starts the TCP server.
func (s *Server) Start() {
	addr := s.Listener.Addr().String()

	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Temporary() {
				continue
			}

			if !nets.IsClosed(err) {
				log.Error("fail to accept the new connection",
					"listenaddr", addr, "err", err)
			}

			s.Handler.OnServerExit(err)
			return
		}

		if s.TLSConfig != nil {
			conn = &tlsConn{
				Conn:   conn,
				first:  true,
				force:  s.ForceTLS,
				config: s.TLSConfig,
			}
		}

		s.Handler.OnConnection(conn)
	}
}

type tlsConn struct {
	config *tls.Config
	force  bool
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
			log.Error("fail to read the first byte from the conneciton",
				"remoteaddr", c.RemoteAddr().String(), "err", c.err)
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
		} else if c.force {
			log.Error("not support the not-TLS conneciton",
				"remoteaddr", c.RemoteAddr().String())

			c.err = errNotSupportNotTLS
			c.Close()
			return
		}
	}
}

var errNotSupportNotTLS = errors.New("not support the not-TLS connection")

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
