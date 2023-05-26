// Copyright 2022~2023 xgfone
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

	"github.com/xgfone/go-apiserver/nets/stream"
	"github.com/xgfone/go-atomicvalue"
	"github.com/xgfone/go-generics/maps"
)

// DefaultManager is the default global middleware manager.
var DefaultManager = NewManager(nil)

// Manager is used to manage a group of the common middlewares,
// which has also implemented the interface http.Handler and
// stream.Handler to be used as a HTTP or stream Handler.
type Manager struct {
	origHandler atomicvalue.Value[interface{}]
	wrapHandler atomicvalue.Value[interface{}]

	lock sync.RWMutex
	maps map[string]Middleware
	mdws atomicvalue.Value[Middlewares]
}

// NewManager returns a new middleware manager.
func NewManager(handler interface{}) *Manager {
	return &Manager{
		maps: make(map[string]Middleware, 8),
		mdws: atomicvalue.NewValue[Middlewares](nil),

		origHandler: atomicvalue.NewValue(handler),
		wrapHandler: atomicvalue.NewValue(handler),
	}
}

func (m *Manager) updateHandler(handler interface{}) {
	if handler == nil {
		m.wrapHandler.Store(nil)
	} else {
		m.wrapHandler.Store(m.GetMiddlewares().Handler(handler))
	}
}

func (m *Manager) updateMiddlewares() {
	mdws := Middlewares(maps.Values(m.maps))
	sort.Stable(mdws)
	m.mdws.Store(mdws)
	if handler := m.GetHandler(); handler == nil {
		m.wrapHandler.Store(nil)
	} else {
		m.wrapHandler.Store(mdws.Handler(handler))
	}
}

// SwapHandler stores the new handler and returns the old.
func (m *Manager) SwapHandler(new interface{}) (old interface{}) {
	m.updateHandler(new)
	return m.origHandler.Swap(new)
}

// SetHandler resets the handler.
func (m *Manager) SetHandler(handler interface{}) {
	m.updateHandler(handler)
	m.origHandler.Store(handler)
}

// GetHandler returns the handler.
func (m *Manager) GetHandler() interface{} {
	return m.origHandler.Load()
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
	maps.Clear(m.maps)
	maps.AddSliceAsValue(m.maps, mws, func(mw Middleware) string { return mw.Name() })
	m.updateMiddlewares()
	m.lock.Unlock()
}

// UpsertMiddlewares adds or updates the middlewares.
func (m *Manager) UpsertMiddlewares(mws ...Middleware) {
	if len(mws) == 0 {
		return
	}

	m.lock.Lock()
	maps.AddSliceAsValue(m.maps, mws, func(mw Middleware) string { return mw.Name() })
	m.updateMiddlewares()
	m.lock.Unlock()
}

// AddMiddleware adds the middleware.
func (m *Manager) AddMiddleware(mw Middleware) (err error) {
	name := mw.Name()
	m.lock.Lock()
	if maps.Add(m.maps, name, mw) {
		m.updateMiddlewares()
	} else {
		err = fmt.Errorf("the middleware named '%s' has existed", name)
	}
	m.lock.Unlock()
	return
}

// DelMiddleware removes and returns the middleware by the name.
//
// If the middleware does not exist, do nothing and return nil.
func (m *Manager) DelMiddleware(name string) Middleware {
	m.lock.Lock()
	mw, ok := maps.Pop(m.maps, name)
	if ok {
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
	return m.mdws.Load()
}

// WrapHandler uses the inner middlewares to wrap the given handler.
func (m *Manager) WrapHandler(handler interface{}) (wrappedHandler interface{}) {
	return m.GetMiddlewares().Handler(handler)
}

// WrappedHandler returns the handler that is wrapped and handled
// by the middlewares.
func (m *Manager) WrappedHandler() interface{} { return m.getHandler() }

func (m *Manager) getHandler() interface{} { return m.wrapHandler.Load() }

// OnConnection implements the interface http.Handler.
//
// Notice: the handler must be a http.Handler.
func (m *Manager) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	m.getHandler().(http.Handler).ServeHTTP(rw, r)
}

// OnConnection implements the interface stream.Handler.
//
// Notice: the handler must be a stream.Handler.
func (m *Manager) OnConnection(c net.Conn) {
	m.getHandler().(stream.Handler).OnConnection(c)
}

// OnServerExit implements the interface stream.Handler.
//
// Notice: the handler must be a stream.Handler.
func (m *Manager) OnServerExit(err error) {
	m.getHandler().(stream.Handler).OnServerExit(err)
}

// OnShutdown implements the interface stream.Handler.
//
// Notice: the handler must be a stream.Handler.
func (m *Manager) OnShutdown(c context.Context) {
	m.getHandler().(stream.Handler).OnShutdown(c)
}
