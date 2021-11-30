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

	"github.com/xgfone/go-apiserver/helper"
)

// Middleware is the http handler middleware.
type Middleware interface {
	Handler(http.Handler) http.Handler
	Name() string
}

// Middlewares is a group of the http handler middlewares.
type Middlewares []Middleware

// Append appends the middlewares into itself.
func (ms *Middlewares) Append(mwds ...Middleware) {
	*ms = append(*ms, mwds...)
}

// Remove removes the middlewares by the names from itself.
func (ms *Middlewares) Remove(names ...string) {
	nameslen := len(names)
	if nameslen == 0 {
		return
	}

	mslen := len(*ms)
	_len := mslen - nameslen
	if _len < 0 {
		_len = mslen
	}

	mdws := make(Middlewares, 0, _len)
	for i := 0; i < mslen; i++ {
		if mw := (*ms)[i]; !helper.InStrings(mw.Name(), names) {
			mdws = append(mdws, mw)
		}
	}
	*ms = mdws
}

// Clone clones itself to a new one.
func (ms Middlewares) Clone() Middlewares {
	return append(Middlewares{}, ms...)
}

// Handler wraps the http handler with the middlewares and returns a new one.
func (ms Middlewares) Handler(handler http.Handler) http.Handler {
	for _len := len(ms) - 1; _len >= 0; _len-- {
		handler = ms[_len].Handler(handler)
	}
	return handler
}

// Index returns the index position where the middleware named name in ms.
//
// If there is not the middleware, return -1.
func (ms Middlewares) Index(name string) int {
	for _len := len(ms) - 1; _len >= 0; _len-- {
		if ms[_len].Name() == name {
			return _len
		}
	}
	return -1
}

// Contains reports whether the middlewares contains the middleware named name.
func (ms Middlewares) Contains(name string) bool {
	return ms.Index(name) > -1
}

type middleware struct {
	name    string
	handler func(http.Handler) http.Handler
}

func (m middleware) Name() string                        { return m.name }
func (m middleware) Handler(h http.Handler) http.Handler { return m.handler(h) }

// NewMiddleware returns a new HTTP handler middleware.
func NewMiddleware(name string, m func(http.Handler) http.Handler) Middleware {
	return middleware{name: name, handler: m}
}

/// ----------------------------------------------------------------------- ///

// MiddlewareManager is used to manage a group of the http middlewares.
type MiddlewareManager struct {
	mdws map[string]Middleware
	lock sync.RWMutex
}

// NewMiddlewareManager returns a new middleware manager.
func NewMiddlewareManager() *MiddlewareManager {
	return &MiddlewareManager{mdws: make(map[string]Middleware, 8)}
}

// AddMiddleware adds the middleware.
func (m *MiddlewareManager) AddMiddleware(mw Middleware) (err error) {
	name := mw.Name()
	m.lock.Lock()
	if _, ok := m.mdws[name]; ok {
		err = fmt.Errorf("the middleware named '%s' has existed", name)
	} else {
		m.mdws[name] = mw
	}
	m.lock.Unlock()
	return
}

// DelMiddleware removes and returns the middleware by the name.
//
// If the middleware does not exist, do nothing and return nil.
func (m *MiddlewareManager) DelMiddleware(name string) Middleware {
	m.lock.Lock()
	mw, ok := m.mdws[name]
	if ok {
		delete(m.mdws, name)
	}
	m.lock.Unlock()
	return mw
}

// GetMiddleware returns the middleware by the name.
//
// If the middleware does not exist, return nil.
func (m *MiddlewareManager) GetMiddleware(name string) Middleware {
	m.lock.RLock()
	mw := m.mdws[name]
	m.lock.RUnlock()
	return mw
}

// GetMiddlewares returns all the middlewares.
func (m *MiddlewareManager) GetMiddlewares() Middlewares {
	m.lock.RLock()
	mdws := make(Middlewares, 0, len(m.mdws))
	for _, mw := range m.mdws {
		mdws = append(mdws, mw)
	}
	m.lock.RUnlock()
	return mdws
}

/// ----------------------------------------------------------------------- ///

// MiddlewareBuilderFunc is a function to build the http middleware.
type MiddlewareBuilderFunc func(name string, config map[string]interface{}) (Middleware, error)

// MiddlewareBuilder is used to build a new middleware with the middleware config.
type MiddlewareBuilder interface {
	Build(name string, config map[string]interface{}) (Middleware, error)
	Type() string
}

type middlewareBuilder struct {
	new MiddlewareBuilderFunc
	typ string
}

func (b middlewareBuilder) Type() string { return b.typ }
func (b middlewareBuilder) Build(n string, c map[string]interface{}) (Middleware, error) {
	return b.new(n, c)
}

// NewMiddlewareBuilder returns a new http middleware builder.
func NewMiddlewareBuilder(typ string, build MiddlewareBuilderFunc) MiddlewareBuilder {
	return middlewareBuilder{typ: typ, new: build}
}

// MiddlewareBuilderManager is used to manage the middleware builder.
type MiddlewareBuilderManager struct {
	builders map[string]MiddlewareBuilder
	lock     sync.RWMutex
}

// NewMiddlewareBuilderManager returns a new middleware builder manager.
func NewMiddlewareBuilderManager() *MiddlewareBuilderManager {
	return &MiddlewareBuilderManager{
		builders: make(map[string]MiddlewareBuilder, 8),
	}
}

// AddBuilder adds the middleware builder.
func (m *MiddlewareBuilderManager) AddBuilder(b MiddlewareBuilder) (err error) {
	typ := b.Type()
	m.lock.Lock()
	if _, ok := m.builders[typ]; ok {
		err = fmt.Errorf("the middleware builder typed '%s' has existed", typ)
	} else {
		m.builders[typ] = b
	}
	m.lock.Unlock()
	return
}

// DelBuilder removes and returns the middleware builder by the type.
//
// If the middleware builder does not exist, do nothing and return nil.
func (m *MiddlewareBuilderManager) DelBuilder(typ string) MiddlewareBuilder {
	m.lock.Lock()
	builder, ok := m.builders[typ]
	if ok {
		delete(m.builders, typ)
	}
	m.lock.Unlock()
	return builder
}

// GetBuilder returns the middleware builder by the type.
//
// If the middleware builder does not exist, return nil.
func (m *MiddlewareBuilderManager) GetBuilder(typ string) MiddlewareBuilder {
	m.lock.RLock()
	builder := m.builders[typ]
	m.lock.RUnlock()
	return builder
}

// GetBuilders returns all the middleware builders.
func (m *MiddlewareBuilderManager) GetBuilders() []MiddlewareBuilder {
	m.lock.RLock()
	builders := make([]MiddlewareBuilder, 0, len(m.builders))
	for _, builder := range m.builders {
		builders = append(builders, builder)
	}
	m.lock.RUnlock()
	return builders
}

// Build uses the builder typed typ to build a middleware named name
// with the config.
func (m *MiddlewareBuilderManager) Build(builderType, middlewareName string,
	middlewareConfig map[string]interface{}) (Middleware, error) {
	if builder := m.GetBuilder(builderType); builder != nil {
		return builder.Build(middlewareName, middlewareConfig)
	}
	return nil, fmt.Errorf("no the http middleware builder typed '%s'", builderType)
}
