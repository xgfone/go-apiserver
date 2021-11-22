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

// Package router implements a router of the http handler.
package router

import (
	"fmt"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"

	ghttp "github.com/xgfone/go-apiserver/http"
	"github.com/xgfone/go-apiserver/http/matcher"
	"github.com/xgfone/go-apiserver/http/ruler"
)

// Middleware is the http handler middleware.
type Middleware func(http.Handler) http.Handler

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
}

// Use uses the given middlewares to act on the handler.
func (r *Route) Use(mws ...Middleware) {
	for _len := len(mws) - 1; _len >= 0; _len-- {
		r.Handler = mws[_len](r.Handler)
	}
}

// NewRoute returns a new Route.
//
// If priority is ZERO, it is equal to the priority of the matcher.
func NewRoute(name string, priority int, m matcher.Matcher, h http.Handler) Route {
	if name == "" {
		panic("the route name is empty")
	}
	if m == nil {
		panic("the route matcher is nil")
	}
	if h == nil {
		panic("the route handler is nil")
	}
	if priority == 0 {
		priority = m.Priority()
	}
	return Route{Name: name, Priority: priority, Matcher: m, Handler: h}
}

// Routes is a group of Routes.
type Routes []Route

func (rs Routes) Len() int           { return len(rs) }
func (rs Routes) Swap(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }
func (rs Routes) Less(i, j int) bool { return rs[j].Priority < rs[i].Priority }

type routesWrapper struct{ Routes }

// Router is the http router to dispatch the request to the different handlers
// by the route rule matcher.
type Router struct {
	NotFound http.Handler
	builder  *ruler.Builder

	lock   sync.RWMutex
	mws    []Middleware
	origs  map[string]Route
	routes Routes
	router atomic.Value
}

// NewRouter returns a new Router.
func NewRouter() *Router {
	r := &Router{
		NotFound: ghttp.Handler404,
		builder:  ruler.NewBuilder(),
		origs:    make(map[string]Route, 16),
	}
	r.router.Store(routesWrapper{})
	return r
}

// Use adds the http handler middlewars and uses them to act on the route handler
// when adding the route.
func (r *Router) Use(mws ...Middleware) {
	r.lock.Lock()
	r.mws = append(r.mws, mws...)
	r.lock.Unlock()
}

// ServeHTTP implements the interface http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	routes := r.router.Load().(routesWrapper).Routes
	for i, _len := 0, len(routes); i < _len; i++ {
		if nreq, ok := routes[i].Matcher.Match(req); ok {
			routes[i].Handler.ServeHTTP(w, nreq)
			return
		}
	}
	r.NotFound.ServeHTTP(w, req)
}

func (r *Router) updateRoutes() {
	routes := make(Routes, 0, len(r.origs))
	for _, route := range r.origs {
		routes = append(routes, route)
	}
	sort.Stable(routes)
	r.routes = routes
	r.router.Store(routesWrapper{Routes: routes})
}

// GetRoutes returns all the routes.
func (r *Router) GetRoutes() (routes Routes) {
	r.lock.RLock()
	routes = make(Routes, len(r.routes))
	copy(routes, r.routes)
	r.lock.RUnlock()
	return
}

// GetRoute returns the route by the name.
func (r *Router) GetRoute(name string) (route Route, ok bool) {
	r.lock.RLock()
	route, ok = r.origs[name]
	r.lock.RUnlock()
	return
}

// AddRoute adds the new route.
func (r *Router) AddRoute(route Route) (err error) {
	r.lock.Lock()
	if _, ok := r.origs[route.Name]; ok {
		err = fmt.Errorf("the route named '%s' has been added", route.Name)
	} else {
		route.Use(r.mws...)
		r.origs[route.Name] = route
		r.updateRoutes()
	}
	r.lock.Unlock()
	return
}

// DelRoute deletes and returns the route by the name.
func (r *Router) DelRoute(name string) (route Route, ok bool) {
	r.lock.Lock()
	if route, ok = r.origs[name]; ok {
		delete(r.origs, name)
		r.updateRoutes()
	}
	r.lock.Unlock()
	return
}

// DelRoutes deletes all the routes by the names.
func (r *Router) DelRoutes(names ...string) {
	if _len := len(names); _len > 0 {
		r.lock.Lock()
		for i := 0; i < _len; i++ {
			delete(r.origs, names[i])
		}
		r.updateRoutes()
		r.lock.Unlock()
	}
}

// UpdateRoutes updates the given routes, which will add the route
// if it does not exist, or update it to the new.
func (r *Router) UpdateRoutes(routes ...Route) {
	if _len := len(routes); _len > 0 {
		r.lock.Lock()
		for i := 0; i < _len; i++ {
			route := routes[i]
			route.Use(r.mws...)
			r.origs[route.Name] = route
		}
		r.updateRoutes()
		r.lock.Unlock()
	}
}

// AddRuleRoute adds the route based on the rule.
func (r *Router) AddRuleRoute(priority int, name, rule string, h http.Handler) error {
	matcher, err := r.builder.Parse(rule)
	if err != nil {
		return err
	}
	return r.AddRoute(NewRoute(name, priority, matcher, h))
}

// Name returns a route builder with the name.
func (r *Router) Name(name string) RouteBuilder {
	return RouteBuilder{router: r, panic: true}.Name(name)
}

// Rule returns a route builder with the matcher rule.
func (r *Router) Rule(matchRule string) RouteBuilder {
	return RouteBuilder{router: r, panic: true}.Rule(matchRule)
}

// RouteBuilder is used to build the route.
type RouteBuilder struct {
	router   *Router
	name     string
	rule     string
	matcher  matcher.Matcher
	priority int
	panic    bool
}

// SetPanic sets the flag to panic when failing to add the route.
//
// Default: true
func (b RouteBuilder) SetPanic(panic bool) RouteBuilder {
	b.panic = panic
	return b
}

// Name sets the name of the route.
func (b RouteBuilder) Name(name string) RouteBuilder {
	b.name = name
	return b
}

// Priority sets the priority of the route.
func (b RouteBuilder) Priority(priority int) RouteBuilder {
	b.priority = priority
	return b
}

// Rule sets the matcher rule of the route.
func (b RouteBuilder) Rule(rule string) RouteBuilder {
	b.rule = rule
	return b
}

// Match sets the matcher of the route.
func (b RouteBuilder) Match(matchers ...matcher.Matcher) RouteBuilder {
	if len(matchers) > 0 {
		b.matcher = matcher.And(matchers...)
	}
	return b
}

// HandlerFunc adds the route with the handler functions.
func (b RouteBuilder) HandlerFunc(handler http.HandlerFunc) error {
	return b.Handler(handler)
}

// Handler adds the route with the handler.
func (b RouteBuilder) Handler(handler http.Handler) error {
	err := b.addRoute(handler)
	if err != nil && b.panic {
		panic(err)
	}
	return err
}

func (b RouteBuilder) addRoute(handler http.Handler) error {
	rule := b.rule
	if b.matcher != nil {
		rule = b.matcher.String()
	}
	if rule == "" {
		return fmt.Errorf("missing the route matcher")
	}

	name := b.name
	if name == "" {
		name = rule
	}

	if b.matcher != nil {
		return b.router.AddRoute(NewRoute(name, b.priority, b.matcher, handler))
	}

	return b.router.AddRuleRoute(b.priority, name, b.rule, handler)
}
