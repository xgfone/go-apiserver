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
	"errors"
	"fmt"
	"net"
	"syscall"
	"time"
)

// Listen opens a stream listener on the given address.
func Listen(network, addr string) (net.Listener, error) {
	listener, err := net.Listen(network, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s listener on %s: %v", network, addr, err)
	}

	if ln, ok := listener.(*net.TCPListener); ok {
		return tcpListener{ln}, nil
	}
	return listener, nil
}

type tcpListener struct{ *net.TCPListener }

func (ln tcpListener) Accept() (net.Conn, error) {
	conn, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}

	if err := conn.SetKeepAlive(true); err != nil {
		conn.Close()
		return nil, err
	}

	if err := conn.SetKeepAlivePeriod(3 * time.Minute); err != nil {
		// Some systems, such as OpenBSD, have no user-settable per-socket
		// TCP keepalive options.
		if !errors.Is(err, syscall.ENOPROTOOPT) {
			conn.Close()
			return nil, err
		}
	}

	return conn, nil
}

var (
	_ Handler      = &ForwardConnListener{}
	_ net.Listener = &ForwardConnListener{}
)

// ForwardConnListener is a listener implementing the interface Handler,
// which accepts and returns a received connection.
type ForwardConnListener struct {
	// OnStop is the callback function when stopping the listener,
	// which may be used to wait all the connections are closed.
	//
	// Default: nil
	OnStop func()

	addr   net.Addr
	connch chan net.Conn
	errch  chan error
}

// NewForwardConnListener returns a new ForwardConnListener.
//
// If connCacheSize is less than 0, use 128.
func NewForwardConnListener(localAddr net.Addr, connCacheSize int) *ForwardConnListener {
	if connCacheSize < 0 {
		connCacheSize = 128
	}

	return &ForwardConnListener{
		connch: make(chan net.Conn, connCacheSize),
		errch:  make(chan error, 1),
		addr:   localAddr,
	}
}

// OnConnection implements the interface Handler.
func (l *ForwardConnListener) OnConnection(conn net.Conn) { l.connch <- conn }

// OnServerExit implements the interface Handler to stop the listener.
func (l *ForwardConnListener) OnServerExit(err error) { l.errch <- err; l.Stop() }

// Stop call the function OnStop to stop the listener.
func (l *ForwardConnListener) Stop() {
	if l.OnStop != nil {
		l.OnStop()
	}
}

// Pending returns the number of the connection in the cache channel.
func (l *ForwardConnListener) Pending() int { return len(l.connch) }

// Addr implements the interface net.Listener.
func (l *ForwardConnListener) Addr() net.Addr { return l.addr }

// Close implements the interface net.Listener, which does nothing.
func (l *ForwardConnListener) Close() (err error) { return }

// Accept implements the interface net.Listener.
func (l *ForwardConnListener) Accept() (conn net.Conn, err error) {
	select {
	case err = <-l.errch:
	case conn = <-l.connch:
	}
	return
}