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
	"sort"
	"sync"
	"sync/atomic"

	"github.com/xgfone/go-apiserver/tcp"
)

type middlewaresWrapper struct{ Middlewares }

// Manager is used to manage a group of the tcp middlewares,
// which has also implemented the interface tcp.Handler
// to be used as a TCP Handler.
type Manager struct {
	orig    tcp.SwitchHandler
	handler tcp.SwitchHandler

	maps map[string]Middleware
	lock sync.RWMutex
	mdws atomic.Value
}

// NewManager returns a new middleware manager.
func NewManager(handler tcp.Handler) *Manager {
	m := &Manager{maps: make(map[string]Middleware, 8)}
	m.orig.Set(handler)
	m.updateMiddlewares()
	return m
}

func (m *Manager) updateHandler(handler tcp.Handler) {
	m.handler.Set(m.GetMiddlewares().Handler(handler))
}

func (m *Manager) updateMiddlewares() {
	mdws := make(Middlewares, 0, len(m.maps))
	for _, mw := range m.maps {
		mdws = append(mdws, mw)
	}

	sort.Stable(mdws)
	m.mdws.Store(middlewaresWrapper{mdws})
	m.handler.Set(mdws.Handler(m.GetHandler()))
}

// SwapHandler stores the new handler and returns the old.
func (m *Manager) SwapHandler(new tcp.Handler) (old tcp.Handler) {
	m.updateHandler(new)
	return m.orig.Swap(new)
}

// SetHandler resets the tcp handler.
func (m *Manager) SetHandler(handler tcp.Handler) {
	m.updateHandler(handler)
	m.orig.Set(handler)
}

// GetHandler returns the tcp handler.
func (m *Manager) GetHandler() tcp.Handler { return m.orig.Get() }

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
func (m *Manager) WrapHandler(handler tcp.Handler) tcp.Handler {
	return m.GetMiddlewares().Handler(handler)
}

// OnConnection implements the interface Handler.
func (m *Manager) OnConnection(c net.Conn) { m.handler.OnConnection(c) }

// OnServerExit implements the interface Handler.
func (m *Manager) OnServerExit(err error) { m.handler.OnServerExit(err) }

// OnShutdown implements the interface Handler.
func (m *Manager) OnShutdown(c context.Context) { m.handler.OnShutdown(c) }
