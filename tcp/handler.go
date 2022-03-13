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
	"net"

	"github.com/xgfone/go-apiserver/internal/atomic"
	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/nets"
)

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

/* ------------------------------------------------------------------------- */

type handlerWrapper struct{ Handler }

var _ Handler = &SwitchHandler{}

// SwitchHandler is a switch handler, which can switch the handler to any one.
type SwitchHandler struct{ handler atomic.Value }

// NewSwitchHandler returns a new switch handler with the initial handler.
func NewSwitchHandler(handler Handler) *SwitchHandler {
	if handler == nil {
		panic("SwitchHandler: the tcp handler is nil")
	}

	sh := new(SwitchHandler)
	sh.Set(handler)
	return sh
}

// Get returns the current handler.
func (sh *SwitchHandler) Get() Handler {
	return sh.handler.Load().(handlerWrapper).Handler
}

// Set sets the tcp handler to new.
func (sh *SwitchHandler) Set(new Handler) {
	sh.handler.Store(handlerWrapper{new})
}

// Swap stores the new handler and returns the old.
func (sh *SwitchHandler) Swap(new Handler) (old Handler) {
	if new == nil {
		panic("SwitchHandler: the new handler is nil")
	}
	return sh.handler.Swap(handlerWrapper{new}).(handlerWrapper).Handler
}

// OnConnection implements the interface Handler, which will forward the call
// to the inner handler.
func (sh *SwitchHandler) OnConnection(c net.Conn) { sh.Get().OnConnection(c) }

// OnServerExit implements the interface Handler, which will forward the call
// to the inner handler.
func (sh *SwitchHandler) OnServerExit(err error) { sh.Get().OnServerExit(err) }

// OnShutdown implements the interface Handler, which will forward the call
// to the inner handler.
func (sh *SwitchHandler) OnShutdown(c context.Context) { sh.Get().OnShutdown(c) }

/* ------------------------------------------------------------------------- */

// IPWhitelistHandler is a tcp handler to support filter the clients
// whose ip is not in the ip whitelist.
type IPWhitelistHandler struct {
	nets.IPChecker
	Handler
}

// NewIPWhitelistHandler returns a new IPWhitelistHandler.
func NewIPWhitelistHandler(handler Handler, ipChecker nets.IPChecker) IPWhitelistHandler {
	if handler == nil {
		panic("IPWhitelistHandler: the tcp handler is nil")
	}
	if ipChecker == nil {
		panic("IPWhitelistHandler: the ip checker is nil")
	}
	return IPWhitelistHandler{Handler: handler, IPChecker: ipChecker}
}

// OnConnection overrides the method OnConnection of the interface Handler.
func (h IPWhitelistHandler) OnConnection(c net.Conn) {
	var ip net.IP
	remoteAddr := c.RemoteAddr()
	switch addr := remoteAddr.(type) {
	case *net.IPAddr:
		ip = addr.IP

	case *net.TCPAddr:
		ip = addr.IP

	case *net.UDPAddr:
		ip = addr.IP

	default:
		host, _ := nets.SplitHostPort(remoteAddr.String())
		ip = net.ParseIP(host)
	}

	if h.CheckIP(ip) {
		h.Handler.OnConnection(c)
	} else {
		c.Close()
		log.Infof("client from '%s' is not allowed", ip)
	}
}
