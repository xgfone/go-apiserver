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
	MiddlewareFunc(requestid.RequestID(nil)),
	MiddlewareFunc(context.Context),
	MiddlewareFunc(logger.Logger),
	MiddlewareFunc(recover.Recover),
}

// Middleware is a http handler middleware.
type Middleware interface {
	Handler(next http.Handler) http.Handler
}

// MiddlewareFunc is the middleware function.
type MiddlewareFunc func(next http.Handler) http.Handler

// Handler implements the interface Middleware.
func (f MiddlewareFunc) Handler(next http.Handler) http.Handler { return f(next) }

// Middlewares is a set of middlewares.
type Middlewares []Middleware

// Clone clones itself to a new middlewares.
func (ms Middlewares) Clone() Middlewares {
	_ms := make(Middlewares, len(ms))
	copy(_ms, ms)
	return _ms
}

// Append appends a set of middlewares and return a new middleware slice.
func (ms Middlewares) Append(m ...Middleware) Middlewares {
	return append(ms, m...)
}

// Handler implements the interface Middleware.
func (ms Middlewares) Handler(next http.Handler) http.Handler {
	for _len := len(ms) - 1; _len >= 0; _len-- {
		next = ms[_len].Handler(next)
	}
	return next
}
