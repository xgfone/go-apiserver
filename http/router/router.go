// Copyright 2021~2022 xgfone
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

	"github.com/xgfone/go-apiserver/http/handler"
	"github.com/xgfone/go-apiserver/middleware"
)

// RouteManager is used to manage the routes.
type RouteManager interface {
	Route(resp http.ResponseWriter, req *http.Request, notFound http.Handler)
}

// Router is the http router to dispatch the request to the different handlers
// by the route rule matcher.
type Router struct {
	// Middlewares is used to manage the middlewares and takes effect
	// before the route manager routes the request.
	Middlewares *middleware.Manager

	// NotFound is used when no route is found.
	NotFound http.Handler

	// RouteManager is used to manage the routes.
	//
	// If implementing the interface RouteManager, use the method Route
	// with NotFound instead of ServeHTTP.
	RouteManager http.Handler
}

// NewRouter returns a new Router with the route manager.
func NewRouter(routeManager http.Handler) *Router {
	r := &Router{
		RouteManager: routeManager,
		Middlewares:  middleware.NewManager(nil),
		NotFound:     handler.Handler404,
	}

	r.Middlewares.SetHandler(http.HandlerFunc(r.serveHTTP))
	return r
}

// ServeHTTP implements the interface http.Handler.
func (r *Router) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	r.Middlewares.ServeHTTP(resp, req)
}

func (r *Router) serveHTTP(resp http.ResponseWriter, req *http.Request) {
	if m, ok := r.RouteManager.(RouteManager); ok {
		m.Route(resp, req, r.NotFound)
	} else {
		r.RouteManager.ServeHTTP(resp, req)
	}
}
