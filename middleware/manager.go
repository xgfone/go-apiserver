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

package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sort"
	"sync"

	"github.com/xgfone/go-apiserver/internal/atomic"
	"github.com/xgfone/go-apiserver/tcp"
)

// DefaultManager is the default global middleware manager.
var DefaultManager = NewManager(nil)

type handlerWrapper struct{ Handler interface{} }
type middlewaresWrapper struct{ Middlewares }

// Manager is used to manage a group of the common middlewares,
// which has also implemented the interface http.Handler and tcp.Handler
// to be used as a HTTP or TCP Handler.
type Manager struct {
	orig    atomic.Value
	handler atomic.Value

	maps map[string]Middleware
	lock sync.RWMutex
	mdws atomic.Value
}

// NewManager returns a new middleware manager.
func NewManager(handler interface{}) *Manager {
	m := &Manager{maps: make(map[string]Middleware, 8)}
	m.orig.Store(handlerWrapper{Handler: handler})
	m.updateMiddlewares()
	return m
}

func (m *Manager) updateHandler(handler interface{}) {
	if handler != nil {
		m.handler.Store(handlerWrapper{m.GetMiddlewares().Handler(handler)})
	}
}

func (m *Manager) updateMiddlewares() {
	mdws := make(Middlewares, 0, len(m.maps))
	for _, mw := range m.maps {
		mdws = append(mdws, mw)
	}

	sort.Stable(mdws)
	m.mdws.Store(middlewaresWrapper{mdws})
	if handler := m.GetHandler(); handler != nil {
		m.handler.Store(handlerWrapper{mdws.Handler(handler)})
	}
}

// SwapHandler stores the new handler and returns the old.
func (m *Manager) SwapHandler(new interface{}) (old interface{}) {
	m.updateHandler(new)
	return m.orig.Swap(new).(handlerWrapper).Handler
}

// SetHandler resets the handler.
func (m *Manager) SetHandler(handler interface{}) {
	m.updateHandler(handler)
	m.orig.Store(handlerWrapper{handler})
}

// GetHandler returns the handler.
func (m *Manager) GetHandler() interface{} {
	return m.orig.Load().(handlerWrapper).Handler
}

// Use is a convenient function to add a group of the given middlewares,
// which will panic with an error when the given middleware has been added.
func (m *Manager) Use(mws ...Middleware) {
	for _, mw := range mws {
		if err := m.AddMiddleware(mw); err != nil {
			panic(err)
		}
	}
}

// Cancel is a convenient function to remove the middlewares by the given names.
func (m *Manager) Cancel(names ...string) {
	for _, name := range names {
		m.DelMiddleware(name)
	}
}

// ResetMiddlewares resets the middlewares.
func (m *Manager) ResetMiddlewares(mws ...Middleware) {
	m.lock.Lock()
	for name := range m.maps {
		delete(m.maps, name)
	}
	for _, mw := range mws {
		m.maps[mw.Name()] = mw
	}
	m.updateMiddlewares()
	m.lock.Unlock()
}

// AddMiddleware adds the middleware.
func (m *Manager) AddMiddleware(mw Middleware) (err error) {
	name := mw.Name()
	m.lock.Lock()
	if _, ok := m.maps[name]; ok {
		err = fmt.Errorf("the middleware named '%s' has existed", name)
	} else {
		m.maps[name] = mw
		m.updateMiddlewares()
	}
	m.lock.Unlock()
	return
}

// DelMiddleware removes and returns the middleware by the name.
//
// If the middleware does not exist, do nothing and return nil.
func (m *Manager) DelMiddleware(name string) Middleware {
	m.lock.Lock()
	mw, ok := m.maps[name]
	if ok {
		delete(m.maps, name)
		m.updateMiddlewares()
	}
	m.lock.Unlock()
	return mw
}

// GetMiddleware returns the middleware by the name.
//
// If the middleware does not exist, return nil.
func (m *Manager) GetMiddleware(name string) Middleware {
	m.lock.RLock()
	mw := m.maps[name]
	m.lock.RUnlock()
	return mw
}

// GetMiddlewares returns all the middlewares.
func (m *Manager) GetMiddlewares() Middlewares {
	return m.mdws.Load().(middlewaresWrapper).Middlewares
}

// WrapHandler uses the inner middlewares to wrap the given handler.
func (m *Manager) WrapHandler(handler interface{}) interface{} {
	return m.GetMiddlewares().Handler(handler)
}

func (m *Manager) getHandler() interface{} {
	return m.handler.Load().(handlerWrapper).Handler
}

// OnConnection implements the interface http.Handler.
//
// Notice: the handler must be a http.Handler.
func (m *Manager) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	m.getHandler().(http.Handler).ServeHTTP(rw, r)
}

// OnConnection implements the interface tcp.Handler.
//
// Notice: the handler must be a tcp.Handler.
func (m *Manager) OnConnection(c net.Conn) {
	m.getHandler().(tcp.Handler).OnConnection(c)
}

// OnServerExit implements the interface tcp.Handler.
//
// Notice: the handler must be a tcp.Handler.
func (m *Manager) OnServerExit(err error) {
	m.getHandler().(tcp.Handler).OnServerExit(err)
}

// OnShutdown implements the interface tcp.Handler.
//
// Notice: the handler must be a tcp.Handler.
func (m *Manager) OnShutdown(c context.Context) {
	m.getHandler().(tcp.Handler).OnShutdown(c)
}
