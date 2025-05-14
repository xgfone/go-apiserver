// Copyright 2023~2025 xgfone
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

	"github.com/xgfone/go-apiserver/http/middleware"
	"github.com/xgfone/go-apiserver/http/reqresp"
)

// Matcher is used to check whether the route matches the request.
type Matcher interface {
	Match(*http.Request) bool
}

// Route is a http request route.
type Route struct {
	handler http.Handler

	// Priority is the priority of the route.
	//
	// The bigger the value, the higher the priority.
	Priority int

	// Matcher is used to match the request.
	Matcher Matcher `json:"-"`

	// Handler is the handler of the route.
	Handler http.Handler `json:"-"`

	// Extra is the extra data of the route.
	Extra any `json:",omitempty"`

	// Desc is the description of the route, which may be matcher string.
	Desc string `json:",omitempty"`

	// Deprecated represents whether the route has been deprecated.
	Deprecated bool `json:",omitempty"`
}

// NewRoute returns a new Route.
func NewRoute(priority int, matcher Matcher, handler http.Handler) Route {
	return Route{Priority: priority, Matcher: matcher, Handler: handler}
}

// WithDesc returns a new Route with the desc.
func (r Route) WithDesc(desc string) Route {
	r.Desc = desc
	return r
}

// WithExtra returns a new Route with the extra.
func (r Route) WithExtra(extra any) Route {
	r.Extra = extra
	return r
}

// ServeHTTP implements the interface http.Handler.
func (r *Route) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if c := reqresp.GetContext(req.Context()); c != nil {
		c.Route = r
	}

	if r.handler != nil {
		r.handler.ServeHTTP(rw, req)
	} else {
		r.Handler.ServeHTTP(rw, req)
	}
}

// Use applies the middlewares on the route handler.
func (r *Route) Use(ms ...middleware.Middleware) {
	handler := r.Handler
	if r.handler != nil {
		handler = r.handler
	}

	r.handler = middleware.Middlewares(ms).Handler(handler)
}
