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
	"sync"
)

// Middleware is the handler middleware.
type Middleware func(Handler) Handler

// Chain returns a new Handler with the middlewares.
func Chain(handler Handler, middlewares ...Middleware) Handler {
	for _len := len(middlewares); _len > 0; _len-- {
		handler = middlewares[_len-1](handler)
	}
	return handler
}

var _ Handler = &MiddlewareHandler{}

// MiddlewareHandler is a handler to support the middleware.
type MiddlewareHandler struct {
	proxy SwitchHandler
	orig  SwitchHandler

	lock sync.RWMutex
	mws  []Middleware
}

// NewMiddlewareHandler returns a new middleware handler.
func NewMiddlewareHandler(handler Handler, mws ...Middleware) *MiddlewareHandler {
	if handler == nil {
		panic("MiddlewareHandler: the tcp handler is nil")
	}

	h := &MiddlewareHandler{orig: newSwitchHandler(handler)}
	h.mws = append([]Middleware{}, mws...)
	h.proxy = newSwitchHandler(Chain(&h.orig, h.mws...))
	return h
}

// Append appends the middlewares and update the handler.
func (h *MiddlewareHandler) Append(mws ...Middleware) *MiddlewareHandler {
	if len(mws) == 0 {
		return h
	}

	h.lock.Lock()
	defer h.lock.Unlock()

	h.mws = append(h.mws, mws...)
	h.proxy.Swap(Chain(&h.orig, h.mws...))

	return h
}

// Get returns the current handler.
func (h *MiddlewareHandler) Get() (handler Handler) { return h.orig.Get() }

// Swap stores the new handler and returns the old.
func (h *MiddlewareHandler) Swap(new Handler) Handler { return h.orig.Swap(new) }

// OnConnection implements the interface Handler.
func (h *MiddlewareHandler) OnConnection(c net.Conn) { h.proxy.OnConnection(c) }

// OnServerExit implements the interface Handler.
func (h *MiddlewareHandler) OnServerExit(err error) { h.proxy.OnServerExit(err) }

// OnShutdown implements the interface Handler.
func (h *MiddlewareHandler) OnShutdown(c context.Context) { h.proxy.OnShutdown(c) }
