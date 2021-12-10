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

// Package handler provides some http handler and middleware functions.
package handler

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/xgfone/go-apiserver/helper"
	"github.com/xgfone/go-apiserver/internal/atomic"
)

// Pre-define some http handlers.
var (
	Handler200 http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	Handler400 http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
	})

	Handler404 http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "close")
		w.WriteHeader(404)
	})
)

/// ----------------------------------------------------------------------- ///

// WrappedHandlerFunc is the function to handle the request with the wrapped handler.
type WrappedHandlerFunc func(http.Handler, http.ResponseWriter, *http.Request)

// WrappedHandler is a http handler which wraps and returns the inner handler.
type WrappedHandler interface {
	WrappedHandler() http.Handler
	http.Handler
}

type wrappedHandler struct {
	Handler http.Handler
	Handle  WrappedHandlerFunc
}

func (wh wrappedHandler) Close() error                 { return helper.Close(wh.Handler) }
func (wh wrappedHandler) WrappedHandler() http.Handler { return wh.Handler }
func (wh wrappedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wh.Handle(wh.Handler, w, r)
}

// WrapHandler returns a new WrappedHandler with the wrapped handler
// and the handling function.
func WrapHandler(handler http.Handler, handlerFunc WrappedHandlerFunc) WrappedHandler {
	return wrappedHandler{Handler: handler, Handle: handlerFunc}
}

// UnwrapHandler unwraps the wrapped innest http handler from handler if it has
// implemented the interface WrappedHandler. Or return the original handler.
func UnwrapHandler(handler http.Handler) http.Handler {
	for {
		if wh, ok := handler.(WrappedHandler); ok {
			handler = wh.WrappedHandler()
		} else {
			break
		}
	}
	return handler
}

/// ----------------------------------------------------------------------- ///

type httpHandlerWrapper struct{ http.Handler }

// SwitchHandler is a HTTP handler that is used to switch the real handler.
type SwitchHandler struct{ handler atomic.Value }

// NewSwitchHandler returns a new switch handler with the initial handler.
func NewSwitchHandler(handler http.Handler) *SwitchHandler {
	if handler == nil {
		panic("SwitchHandler: the http handler is nil")
	}

	sh := &SwitchHandler{}
	sh.handler.Store(httpHandlerWrapper{handler})
	return sh
}

// Close implements the interface io.Closer.
func (sh *SwitchHandler) Close() error { return helper.Close(sh.Get()) }

// Set sets the http handler to new.
func (sh *SwitchHandler) Set(new http.Handler) {
	sh.handler.Store(httpHandlerWrapper{new})
}

// Get returns the current handler.
func (sh *SwitchHandler) Get() http.Handler {
	return sh.handler.Load().(httpHandlerWrapper).Handler
}

// Swap stores the new handler and returns the old.
func (sh *SwitchHandler) Swap(new http.Handler) (old http.Handler) {
	if new == nil {
		panic("SwitchHandler.Swap(): the new http handler is nil")
	}
	return sh.handler.Swap(httpHandlerWrapper{new}).(httpHandlerWrapper).Handler
}

// ServeHTTP implements the interface http.Handler.
func (sh *SwitchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sh.handler.Load().(httpHandlerWrapper).Handler.ServeHTTP(w, r)
}

// WrappedHandler implements the interface WrappedHandler.
func (sh *SwitchHandler) WrappedHandler() http.Handler { return sh.Get() }

/// ----------------------------------------------------------------------- ///

// MiddlewareHandler is the HTTP handler with the middlewares, that's,
// the middlewares will handle the request before the http handler.
type MiddlewareHandler struct {
	handler SwitchHandler
	orig    SwitchHandler

	lock sync.RWMutex
	mdws Middlewares
}

// NewMiddlewareHandler returns a new the HTTP handler based on the middlewares.
func NewMiddlewareHandler(handler http.Handler, mdws ...Middleware) *MiddlewareHandler {
	var mh MiddlewareHandler
	mh.orig.Set(handler)
	mh.Use(mdws...)
	return &mh
}

// Close implements the interface io.Closer.
func (mh *MiddlewareHandler) Close() error { return mh.orig.Close() }

// WrappedHandler implements the interface WrappedHandler.
func (mh *MiddlewareHandler) WrappedHandler() http.Handler { return mh.Get() }

// ServeHTTP implements the interface http.Handler.
func (mh *MiddlewareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mh.handler.ServeHTTP(w, r)
}

// Get returns the original handler.
func (mh *MiddlewareHandler) Get() http.Handler { return mh.orig.Get() }

// Swap stores the new handler as the original handler and returns the old.
func (mh *MiddlewareHandler) Swap(new http.Handler) (old http.Handler) {
	if new == nil {
		panic("MiddlewareHandler.Swap(): the new http handler is nil")
	}

	old = mh.orig.Swap(new)
	mh.lock.RLock()
	defer mh.lock.RUnlock()
	mh.updateHandler()
	return
}

// Handler returns a new http handler which is decorated by the middlewares.
func (mh *MiddlewareHandler) Handler(handler http.Handler) http.Handler {
	mh.lock.Lock()
	defer mh.lock.Unlock()
	return mh.mdws.Handler(handler)
}

// Use appends the http handler middlewars and uses them to the http handler.
func (mh *MiddlewareHandler) Use(mws ...Middleware) {
	mh.lock.Lock()
	defer mh.lock.Unlock()
	mh.mdws.Append(mws...)
	mh.updateHandler()
}

// UseReset is the same as Use, but resets the route middlewares to mws.
func (mh *MiddlewareHandler) UseReset(mws ...Middleware) {
	mh.lock.Lock()
	defer mh.lock.Unlock()
	mh.mdws = append(Middlewares{}, mws...)
	mh.updateHandler()
}

// Unuse removes the http handler middlewares by the names.
func (mh *MiddlewareHandler) Unuse(names ...string) {
	mh.lock.Lock()
	defer mh.lock.Unlock()
	mh.mdws.Remove(names...)
	mh.updateHandler()
}

func (mh *MiddlewareHandler) updateHandler() {
	mh.handler.Set(mh.mdws.Handler(mh.Get()))
}

/// ----------------------------------------------------------------------- ///

// Manager is used to manage the http handler.
type Manager struct {
	lock     sync.RWMutex
	handlers map[string]http.Handler
}

// NewManager returns a new http handler manager.
func NewManager() *Manager {
	return &Manager{handlers: make(map[string]http.Handler, 8)}
}

// AddHandler adds the named http handler.
func (m *Manager) AddHandler(name string, handler http.Handler) (err error) {
	if name == "" {
		return errors.New("the http handler name is empty")
	} else if handler == nil {
		return errors.New("the http handler is nil")
	}

	m.lock.Lock()
	if _, ok := m.handlers[name]; ok {
		err = fmt.Errorf("the http handler namde '%s' has existed", name)
	} else {
		m.handlers[name] = handler
	}
	m.lock.Unlock()

	return
}

// DelHandler deletes the http handler by the name.
//
// If the http handler does not exist, do nothing and return nil.
func (m *Manager) DelHandler(name string) http.Handler {
	m.lock.Lock()
	handler, ok := m.handlers[name]
	if ok {
		delete(m.handlers, name)
	}
	m.lock.Unlock()
	return handler
}

// GetHandler returns the http handler by the name.
//
// If the http handler does not exist, return nil.
func (m *Manager) GetHandler(name string) http.Handler {
	m.lock.RLock()
	handler := m.handlers[name]
	m.lock.RUnlock()
	return handler
}

// GetHandlers returns all the http handlers.
func (m *Manager) GetHandlers() map[string]http.Handler {
	m.lock.RLock()
	handlers := make(map[string]http.Handler, len(m.handlers))
	for name, handler := range m.handlers {
		handlers[name] = handler
	}
	m.lock.RUnlock()
	return handlers
}