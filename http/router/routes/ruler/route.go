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

package ruler

import (
	"net/http"

	"github.com/xgfone/go-apiserver/http/matcher"
	"github.com/xgfone/go-apiserver/http/middleware"
)

/*
// Manager is used to manage a set of routes.
type Manager interface {
	// MatchRoute uses the registered routes to match the http request,
	// and returns the matched route.
	MatchRoute(orig *http.Request) (newreq *http.Request, route Route, ok bool)

	// GetRoute returns the route by the given name.
	GetRoute(name string) (route Route, ok bool)

	// GetRoutes returns all the registered routes.
	GetRoutes() (routes Routes)

	// AddRoute is used to add the given route.
	AddRoute(route Route) error

	// DelRoute deletes and returns the route by the given name.
	DelRoute(name string) (route Route, ok bool)

	// UpdateRoutes updates the given routes, which will add the route
	// if it does not exist, or update it to the new.
	UpdateRoutes(routes ...Route) error

	// ResetRoutes discards all the original routes and resets them to routes.
	ResetRoutes(routes ...Route) error
}
*/

// Routes is a group of Routes.
type Routes []Route

func (rs Routes) Len() int           { return len(rs) }
func (rs Routes) Swap(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }
func (rs Routes) Less(i, j int) bool { return rs[j].Priority < rs[i].Priority }

// Route is a http request route.
type Route struct {
	// Name is the unique name of the route.
	Name string

	// Priority is the priority of the route.
	//
	// The bigger the value, the higher the priority.
	Priority int

	// Matcher is used to match the request.
	Matcher matcher.Matcher

	// Handler is the handler of the route.
	http.Handler

	handler http.Handler
}

// NewRoute returns a new Route.
//
// If priority is ZERO, it is equal to the priority of the matcher.
func NewRoute(name string, priority int, m matcher.Matcher, h http.Handler) Route {
	if priority == 0 {
		priority = m.Priority()
	}
	return Route{Name: name, Priority: priority, Matcher: m, Handler: h}
}

// ServeHTTP implements the interface http.Handler.
func (r *Route) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if r.handler != nil {
		r.handler.ServeHTTP(resp, req)
	} else {
		r.Handler.ServeHTTP(resp, req)
	}
}

// Use uses the given middlewares to act on the handler.
func (r *Route) Use(mws ...middleware.Middleware) {
	if r.Handler == nil {
		panic("the route handler is nil")
	}

	r.handler = r.Handler
	for _len := len(mws) - 1; _len >= 0; _len-- {
		r.handler = mws[_len].Handler(r.handler)
	}
}
