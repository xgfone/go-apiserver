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
	"net"
	"sync"

	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/nets"
	"github.com/xgfone/go-atomicvalue"
)

// Handler is used to handle the stream connection.
type Handler interface {
	// OnConnection is called when a new connection comes.
	//
	// Notice: it is the responsibility of the handler to close the connection.
	OnConnection(net.Conn)

	// OnServerExit is called when the server exits,
	// which should return only when all the connections are closed.
	OnServerExit(err error)
}

/* ------------------------------------------------------------------------- */

var _ Handler = &SwitchHandler{}

// SwitchHandler is a switch handler, which can switch the handler to any one.
type SwitchHandler struct{ handler atomicvalue.Value[Handler] }

// NewSwitchHandler returns a new switch handler with the initial handler.
func NewSwitchHandler(handler Handler) *SwitchHandler {
	if handler == nil {
		panic("SwitchHandler: the stream handler is nil")
	}
	return &SwitchHandler{handler: atomicvalue.NewValue(handler)}
}

// Get returns the current handler.
func (sh *SwitchHandler) Get() Handler { return sh.handler.Load() }

// Set sets the handler to new.
func (sh *SwitchHandler) Set(new Handler) { sh.handler.Store(new) }

// Swap stores the new handler and returns the old.
func (sh *SwitchHandler) Swap(new Handler) (old Handler) {
	if new == nil {
		panic("SwitchHandler.Swap: the new handler is nil")
	}
	return sh.handler.Swap(new)
}

// OnConnection implements the interface Handler, which will forward the call
// to the inner handler.
func (sh *SwitchHandler) OnConnection(c net.Conn) { sh.Get().OnConnection(c) }

// OnServerExit implements the interface Handler, which will forward the call
// to the inner handler.
func (sh *SwitchHandler) OnServerExit(err error) { sh.Get().OnServerExit(err) }

/* ------------------------------------------------------------------------- */

// IPWhitelistHandler is a stream handler to support filter the clients
// whose ip is not in the ip whitelist.
type IPWhitelistHandler struct {
	nets.IPChecker
	Handler
}

// NewIPWhitelistHandler returns a new IPWhitelistHandler.
func NewIPWhitelistHandler(handler Handler, ipChecker nets.IPChecker) IPWhitelistHandler {
	if handler == nil {
		panic("IPWhitelistHandler: the stream handler is nil")
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

/* ------------------------------------------------------------------------- */

// NewFuncHandler returns a new handler, based on the a simple function,
// of the stream server.
//
// When the stream server is stopped, the context is done.
func NewFuncHandler(handler func(context.Context, net.Conn)) Handler {
	c, cancel := context.WithCancel(context.Background())
	return &funcHandler{Handler: handler, cancel: cancel, context: c}
}

type funcHandler struct {
	Handler func(context.Context, net.Conn)

	cancel  func()
	context context.Context
	wg      sync.WaitGroup
}

func (h *funcHandler) onConnection(conn net.Conn) {
	defer h.wg.Done()
	defer conn.Close()
	h.Handler(h.context, conn)
}
func (h *funcHandler) OnConnection(conn net.Conn) {
	h.wg.Add(1)
	go h.onConnection(conn)
}

func (h *funcHandler) OnServerExit(err error) {
	h.cancel()
	h.wg.Wait()
}
