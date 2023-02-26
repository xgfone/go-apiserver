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
	"github.com/xgfone/go-apiserver/middleware"
)

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
	Handler http.Handler `json:"-"`

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
//
// Notice: if using Use to configure the middlewares on the http handler,
// you should use the method ServeHTTP to handle the http request.
func (r *Route) Use(mws ...middleware.Middleware) {
	if r.Handler == nil {
		panic("the route handler is nil")
	}

	r.handler = r.Handler
	for _len := len(mws) - 1; _len >= 0; _len-- {
		r.handler = mws[_len].Handler(r.handler).(http.Handler)
	}
}
