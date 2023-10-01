// Copyright 2023 xgfone
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

// Package middleware defines a http handler middleware.
package middleware

import (
	"net/http"
	"slices"

	"github.com/xgfone/go-apiserver/helper"
	"github.com/xgfone/go-apiserver/http/middleware/context"
	"github.com/xgfone/go-apiserver/http/middleware/logger"
	"github.com/xgfone/go-apiserver/http/middleware/recover"
	"github.com/xgfone/go-apiserver/http/middleware/requestid"
)

var (
	_ Middleware = MiddlewareFunc(nil)
	_ Middleware = Middlewares(nil)
)

// DefaultMiddlewares is a set of the default middlewares.
var DefaultMiddlewares = Middlewares{
	New("requestid", 10, requestid.RequestID(nil)),
	New("context", 20, context.Context),
	New("logger", 30, logger.Logger),
	New("recover", 40, recover.Recover),
}

// Middleware is a http handler middleware.
type Middleware interface {
	Handler(next http.Handler) http.Handler
}

// MiddlewareFunc is the middleware function.
type MiddlewareFunc func(next http.Handler) http.Handler

// Handler implements the interface Middleware.
func (f MiddlewareFunc) Handler(next http.Handler) http.Handler { return f(next) }

func funcs2mws(fs []MiddlewareFunc) Middlewares {
	if len(fs) == 0 {
		return nil
	}

	ms := make(Middlewares, len(fs))
	for i, f := range fs {
		ms[i] = f
	}
	return ms
}

// Middlewares is a set of middlewares.
type Middlewares []Middleware

// Clone clones itself to a new middlewares.
func (ms Middlewares) Clone() Middlewares {
	return slices.Clone(ms)
}

// InsertFunc inserts a set of function middlewares into the front
// and return a new middleware slice.
func (ms Middlewares) InsertFunc(m ...MiddlewareFunc) Middlewares {
	switch _len := len(m); _len {
	case 0:
		return ms

	case 1:
		return ms.Insert(m[0])

	case 2:
		return ms.Insert(m[0], m[1])

	default:
		return ms.Insert(funcs2mws(m)...)
	}
}

// AppendFunc appends a set of function middlewares
// and return a new middleware slice.
func (ms Middlewares) AppendFunc(m ...MiddlewareFunc) Middlewares {
	switch _len := len(m); _len {
	case 0:
		return ms

	case 1:
		return ms.Append(m[0])

	case 2:
		return ms.Append(m[0], m[1])

	default:
		return ms.Append(funcs2mws(m)...)
	}
}

// Insert inserts a set of middlewares into the front
// and return a new middleware slice.
func (ms Middlewares) Insert(m ...Middleware) Middlewares {
	return helper.MergeSlice(m, ms)
}

// Append appends a set of middlewares and return a new middleware slice.
func (ms Middlewares) Append(m ...Middleware) Middlewares {
	return helper.MergeSlice(ms, m)
}

// Handler implements the interface Middleware.
func (ms Middlewares) Handler(next http.Handler) http.Handler {
	for _len := len(ms) - 1; _len >= 0; _len-- {
		next = ms[_len].Handler(next)
	}
	return next
}

// middleware is a named priority middleware.
type middleware struct {
	f MiddlewareFunc
	n string
	p int
}

// New returns a new named priority middleware,
// which has the methods as follows:
//
//	Name() string
//	Priority() int
//
// For the priority, the smaller the value, the higher the priority.
func New(name string, priority int, mfunc MiddlewareFunc) Middleware {
	if mfunc == nil {
		panic("Middleware.New: the middleware function must not be nil")
	}
	return &middleware{n: name, p: priority, f: mfunc}
}

func (m *middleware) Name() string                           { return m.n }
func (m *middleware) Priority() int                          { return m.p }
func (m *middleware) Handler(next http.Handler) http.Handler { return m.f(next) }

// Sort sorts a set of middlewares, which must have implemented
//
//	interface{ Priority() int }
func Sort(ms []Middleware) {
	slices.SortStableFunc(ms, func(a, b Middleware) int {
		return getPriority(a) - getPriority(b)
	})
}

func getPriority(m Middleware) int {
	return m.(interface{ Priority() int }).Priority()
}
