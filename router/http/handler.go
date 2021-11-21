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

package http

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/xgfone/go-apiserver/internal/atomic"
)

// Pre-define some http handlers.
var (
	Handler200 = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	Handler400 = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
	})

	Handler404 = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "close")
		w.WriteHeader(404)
	})
)

// Middleware is the http handler middleware.
type Middleware func(http.Handler) http.Handler

/// ----------------------------------------------------------------------- ///

// NamedHandler is the named http handler.
type NamedHandler interface {
	Name() string
	http.Handler
}

type namedHandler struct {
	http.Handler
	name string
}

func (h namedHandler) Name() string { return h.name }

// NewNamedHandler returns a new named http handler.
func NewNamedHandler(name string, handler http.Handler) NamedHandler {
	return namedHandler{name: name, Handler: handler}
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

/// ----------------------------------------------------------------------- ///

// HandlerManager is used to manage the http handler.
type HandlerManager struct {
	lock     sync.RWMutex
	handlers map[string]http.Handler
}

// NewHandlerManager returns a new http handler manager.
func NewHandlerManager() *HandlerManager {
	return &HandlerManager{handlers: make(map[string]http.Handler, 8)}
}

// AddHandler adds the named http handler.
func (m *HandlerManager) AddHandler(name string, handler http.Handler) (err error) {
	if name == "" {
		panic("the http handler name is empty")
	} else if handler == nil {
		panic("the http handler is nil")
	}

	m.lock.Lock()
	if _, ok := m.handlers[name]; ok {
		err = fmt.Errorf("the http handler namde '%s' has existed", name)
	} else {
		m.handlers[name] = handler
	}

	return
}

// DelHandler deletes the http handler by the name.
//
// If the http handler does not exist, do nothing and return nil.
func (m *HandlerManager) DelHandler(name string) http.Handler {
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
func (m *HandlerManager) GetHandler(name string) http.Handler {
	m.lock.RLock()
	handler := m.handlers[name]
	m.lock.RUnlock()
	return handler
}

// GetHandlers returns all the http handlers.
func (m *HandlerManager) GetHandlers() map[string]http.Handler {
	m.lock.RLock()
	handlers := make(map[string]http.Handler, len(m.handlers))
	for name, handler := range m.handlers {
		handlers[name] = handler
	}
	m.lock.RUnlock()
	return handlers
}
