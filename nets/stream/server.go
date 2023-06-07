// Copyright 2021~2023 xgfone
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

package stream

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/nets"
)

var defaultCipherSuites []uint16

func init() {
	for _, cs := range tls.CipherSuites() {
		defaultCipherSuites = append(defaultCipherSuites, cs.ID)
	}
}

func getCipherSuites() []uint16 {
	if len(defaultCipherSuites) == 0 {
		return nil
	}
	return append([]uint16{}, defaultCipherSuites...)
}

type tlsOption struct {
	TLSConfig *tls.Config
	ForceTLS  bool
}

// Server implements a server based on the stream.
type Server struct {
	Handler  Handler
	Listener net.Listener
	tlsconf  atomic.Value

	stopped int32
	stops   []func()
}

// NewServer returns a new Server.
func NewServer(ln net.Listener, handler Handler) *Server {
	server := &Server{Listener: ln, Handler: handler}
	server.SetTLSConfig(nil, false)
	return server
}

// SetTLSConfig sets the TLS configuration, which is thread-safe.
//
// If tlsConfig is set and forceTLS is true, the client must use TLS.
// If tlsConfig is set and forceTLS is false, the client maybe use TLS or not-TLS.
// If tlsConfig is not set, forceTLS is ignored and the client must use not-TLS.
func (s *Server) SetTLSConfig(tlsConfig *tls.Config, forceTLS bool) {
	if tlsConfig != nil && len(tlsConfig.CipherSuites) == 0 {
		tlsConfig.CipherSuites = getCipherSuites()
	}
	s.tlsconf.Store(tlsOption{TLSConfig: tlsConfig, ForceTLS: forceTLS})
}

// GetTLSConfig returns the TLS configuration, which is thread-safe.
func (s *Server) GetTLSConfig() (tlsConfig *tls.Config, forceTLS bool) {
	opt := s.tlsconf.Load().(tlsOption)
	return opt.TLSConfig, opt.ForceTLS
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

// Start starts the stream server.
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

		if tlsConfig, forceTLS := s.GetTLSConfig(); tlsConfig != nil {
			if forceTLS {
				conn = tls.Server(conn, tlsConfig)
			} else {
				conn = &TryTLSConn{Conn: conn, Config: tlsConfig}
			}
		}

		s.Handler.OnConnection(conn)
	}
}

// TryTLSConn is used to try to ensure the TLS connection
// if the client enables TLS.
type TryTLSConn struct {
	*tls.Config
	net.Conn

	check bool
}

func (c *TryTLSConn) ensure() (err error) {
	if !c.check {
		c.check = true
		c.Conn, err = checkTLS(c.Conn, c.Config, time.Second*2)
	}
	return
}

// Read implements the interface io.Reader to override the Read method.
func (c *TryTLSConn) Read(b []byte) (n int, err error) {
	if err = c.ensure(); err == nil {
		n, err = c.Conn.Read(b)
	}
	return
}

// GetConn returns the ensured connection.
func (c *TryTLSConn) GetConn() (conn net.Conn, err error) {
	err = c.ensure()
	conn = c.Conn
	return
}

func checkTLS(conn net.Conn, config *tls.Config, timeout time.Duration) (net.Conn, error) {
	if timeout > 0 {
		conn.SetReadDeadline(time.Now().Add(timeout))
	}

	var bs [1]byte
	if n, err := conn.Read(bs[:]); err != nil {
		conn.Close()
		if !errors.Is(err, syscall.ECONNRESET) && !strings.HasSuffix(err.Error(), "connection reset by peer") {
			log.Error("fail to read the first byte from the conneciton",
				"remoteaddr", conn.RemoteAddr().String(), "err", err)
		}
		return conn, err
	} else if n == 0 {
		conn.Close()
		log.Error("read the zero byte from the connection", "remoteaddr", conn.RemoteAddr().String())
		return conn, io.EOF
	}

	if timeout > 0 {
		conn.SetReadDeadline(time.Time{})
	}

	conn = &peekedConn{Conn: conn, Peeked: int16(bs[0])}

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
		conn = tls.Server(conn, config)
	}

	return conn, nil
}

type peekedConn struct {
	Peeked int16
	net.Conn
}

func (c *peekedConn) Read(p []byte) (n int, err error) {
	if c.Peeked == -1 {
		return c.Conn.Read(p)
	}

	p[0], c.Peeked = byte(c.Peeked), -1
	if len(p) == 1 {
		return 1, nil
	}

	p = p[1:]
	n, err = c.Conn.Read(p)
	n++

	return
}
