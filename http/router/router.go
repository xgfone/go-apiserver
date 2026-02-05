// Copyright 2021~2023 xgfone
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

// Package router implements a router of the http handler.
package router

import (
	"net/http"

	"github.com/xgfone/go-apiserver/http/middleware"
	"github.com/xgfone/go-apiserver/http/router/ruler"
)

// DefaultRouter is the default global router.
var DefaultRouter = NewDefaultRouter(ruler.DefaultRouter)

// Router is a http router that manages all the http middlewares uniformly.
type Router struct {
	// Middlewares is used to manage the middlewares and takes effect
	// before routing the request.
	Middlewares *middleware.Manager

	// Router is used to manage the routes.
	Router http.Handler
}

// NewRouter returns a new Router with the next router handler.
func NewRouter(router http.Handler) *Router {
	r := &Router{Router: router}
	r.Middlewares = middleware.NewManager(http.HandlerFunc(r.serveHTTP))
	return r
}

// NewDefaultRouter is the same as NewRouter, but also adds the default
// middlewares, that's middleware.DefaultMiddlewares.
func NewDefaultRouter(router http.Handler) *Router {
	r := NewRouter(router)
	r.Middlewares.Append(middleware.DefaultMiddlewares...)
	return r
}

// Use appends a set of middlewares into the middleware manager.
func (r *Router) Use(ms ...middleware.Middleware) *Router {
	r.Middlewares.Append(ms...)
	return r
}

// UseFunc appends a set of function middlewares into the middleware manager.
func (r *Router) UseFunc(ms ...middleware.MiddlewareFunc) *Router {
	r.Middlewares.AppendFunc(ms...)
	return r
}

// UseContextHandler appends the context handler middlewares that act on the later handler.
func (r *Router) UseContextHandler(handlers ...middleware.ContextHandler) *Router {
	r.Middlewares.AppendContextHandler(handlers...)
	return r
}

// ServeHTTP implements the interface http.Handler.
func (r *Router) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	r.Middlewares.ServeHTTP(resp, req)
}

func (r *Router) serveHTTP(resp http.ResponseWriter, req *http.Request) {
	r.Router.ServeHTTP(resp, req)
}
