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

// Package ruler implements a route manager based on the ruler.
package ruler

import (
	"net/http"
	"slices"

	"github.com/xgfone/go-apiserver/http/middleware"
	"github.com/xgfone/go-toolkit/httpx"
)

// DefaultRouter is the default router.
var DefaultRouter = NewRouter()

// Router is used to manage a set of routes based on the ruler.
type Router struct {
	// NotFound is used when the manager is used as http.Handler
	// and does not find the route.
	//
	// Default: httpx.Handler404
	NotFound http.Handler

	// Middlewares is used to manage the middlewares and applied to each route
	// when registering it. So, the middlewares will be run after routing
	// and never be run if not found the route.
	Middlewares *middleware.Manager

	routes []Route
}

// NewRouter returns a new router.
func NewRouter() *Router {
	return &Router{Middlewares: middleware.NewManager(nil)}
}

// ServeHTTP implements the interface http.Handler.
func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for i, _len := 0, len(r.routes); i < _len; i++ {
		route := &r.routes[i]
		if route.Matcher.Match(req) {
			route.ServeHTTP(rw, req)
			return
		}
	}

	if r.NotFound != nil {
		r.NotFound.ServeHTTP(rw, req)
	} else {
		httpx.Handler404.ServeHTTP(rw, req)
	}
}

// Routes returns all the registered routes, which must be read-only.
func (r *Router) Routes() (routes []Route) { return r.routes }

// Register registers the route.
//
// NOTICE: if both routes match a request, the handler of the higher priority
// route will be executed, and that of the lower never be executed.
func (r *Router) Register(route Route) {
	r.checkRoute(&route)
	r.routes = append(r.routes, route)
	r.updateRoutes()
}

func (r *Router) checkRoute(route *Route) {
	if route.Matcher == nil {
		panic("the route matcher must not be empty")
	}
	if route.Handler == nil {
		panic("the route handler must not be empty")
	}
	route.Use(r.Middlewares.Middlewares())
}

func (r *Router) updateRoutes() {
	slices.SortFunc(r.routes, func(a, b Route) int {
		return b.Priority - a.Priority
	})
}
