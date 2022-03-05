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
	"errors"
	"fmt"
	"net"
	"runtime"
	"syscall"
	"time"
)

// Listen opens a tcp listener on the given address.
func Listen(addr string) (net.Listener, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to open listener on %s: %v", addr, err)
	}
	return keepAliveListener{listener.(*net.TCPListener)}, nil
}

type keepAliveListener struct{ *net.TCPListener }

func (ln keepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}

	if err := tc.SetKeepAlive(true); err != nil {
		return nil, err
	}

	if err := tc.SetKeepAlivePeriod(3 * time.Minute); err != nil {
		// Some systems, such as OpenBSD, have no user-settable per-socket TCP
		// keepalive options.
		if !errors.Is(err, syscall.ENOPROTOOPT) {
			return nil, err
		}
	}

	return tc, nil
}

var (
	_ Handler      = &ForwardConnListener{}
	_ net.Listener = &ForwardConnListener{}
)

// ForwardConnListenerConfig is used to
type ForwardConnListenerConfig struct {
	// Close is the callback function when closing the listener.
	Close func() error

	// OnShutdown is the callback function when calling the method OnShutdown.
	OnShutdown func(context.Context)

	// ConnCacheSize is the size of the connection cache.
	//
	// Default: runtime.NumCPU()
	ConnCacheSize int
}

// ForwardConnListener is a listener implementing the interface Handler,
// which accepts and returns a received connection.
type ForwardConnListener struct {
	onShutdown func(context.Context)
	onClose    func() error

	addr   net.Addr
	connch chan net.Conn
	errch  chan error
}

// NewForwardConnListener returns a new ForwardConnListener.
func NewForwardConnListener(localAddr net.Addr, c *ForwardConnListenerConfig) *ForwardConnListener {
	var config ForwardConnListenerConfig
	if c != nil {
		config = *c
	}

	if config.ConnCacheSize < 0 {
		config.ConnCacheSize = 0
	} else if config.ConnCacheSize == 0 {
		config.ConnCacheSize = runtime.NumCPU()
	}

	return &ForwardConnListener{
		onShutdown: config.OnShutdown,
		onClose:    config.Close,
		connch:     make(chan net.Conn, config.ConnCacheSize),
		errch:      make(chan error),
		addr:       localAddr,
	}
}

// OnConnection implements the interface Handler.
func (l *ForwardConnListener) OnConnection(conn net.Conn) { l.connch <- conn }

// OnServerExit implements the interface Handler.
func (l *ForwardConnListener) OnServerExit(err error) { l.errch <- err }

// OnShutdown implements the interface Handler.
func (l *ForwardConnListener) OnShutdown(ctx context.Context) {
	if l.onShutdown != nil {
		l.onShutdown(ctx)
	}
}

// Addr implements the interface net.Listener.
func (l *ForwardConnListener) Addr() net.Addr { return l.addr }

// Close implements the interface net.Listener.
func (l *ForwardConnListener) Close() (err error) {
	if l.onClose != nil {
		err = l.onClose()
	}
	return
}

// Accept implements the interface net.Listener.
func (l *ForwardConnListener) Accept() (conn net.Conn, err error) {
	select {
	case conn = <-l.connch:
	case err = <-l.errch:
	}
	return
}
