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

// Package middleware provides the common middleware functions for the handler.
package middleware

// Middleware is the common handler middleware.
type Middleware interface {
	// Name returns the name of the middleware.
	Name() string

	// The smaller the value, the higher the priority and the middleware
	// is executed preferentially.
	Priority() int

	// Handler is used to wrap the handler and returns a new one.
	Handler(wrappedhandler interface{}) (newHandler interface{})
}

type middleware struct {
	name     string
	priority int
	handler  func(interface{}) interface{}
}

func (m middleware) Name() string                      { return m.name }
func (m middleware) Priority() int                     { return m.priority }
func (m middleware) Handler(h interface{}) interface{} { return m.handler(h) }

// New is short for NewMiddleware.
func New(name string, prio int, f func(h interface{}) interface{}) Middleware {
	return NewMiddleware(name, prio, f)
}

// NewMiddleware returns a new common handler middleware.
func NewMiddleware(name string, prio int, f func(h interface{}) interface{}) Middleware {
	return middleware{name: name, priority: prio, handler: f}
}

// Middlewares is a group of the common handler middlewares.
type Middlewares []Middleware

func (ms Middlewares) Len() int           { return len(ms) }
func (ms Middlewares) Swap(i, j int)      { ms[i], ms[j] = ms[j], ms[i] }
func (ms Middlewares) Less(i, j int) bool { return ms[i].Priority() < ms[j].Priority() }

// Handler wraps the handler with the middlewares and returns a new one.
func (ms Middlewares) Handler(handler interface{}) interface{} {
	for _len := len(ms) - 1; _len >= 0; _len-- {
		handler = ms[_len].Handler(handler)
	}
	return handler
}

// Clone clones itself and appends the new middlewares to the new.
func (ms Middlewares) Clone(news ...Middleware) Middlewares {
	mws := make(Middlewares, len(ms)+len(news))
	copy(mws, ms)
	copy(mws[len(ms):], news)
	return mws
}
