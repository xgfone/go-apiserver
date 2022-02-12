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
	"errors"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/xgfone/go-apiserver/http/handler"
	"github.com/xgfone/go-apiserver/http/matcher"
)

// Middleware is the http handler middleware.
type Middleware = handler.Middleware

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

// Use uses the given middlewares to act on the handler.
func (r *Route) Use(mws ...Middleware) {
	if r.Handler == nil {
		panic("the route handler is nil")
	}

	r.handler = r.Handler
	for _len := len(mws) - 1; _len >= 0; _len-- {
		r.handler = mws[_len].Handler(r.handler)
	}
}

// NewRoute is the same as NewRouteWithError, but panics if returning an error.
func NewRoute(name string, priority int, m matcher.Matcher, h http.Handler) Route {
	route, err := NewRouteWithError(name, priority, m, h)
	if err != nil {
		panic(err)
	}
	return route
}

// NewRouteWithError returns a new Route.
//
// If priority is ZERO, it is equal to the priority of the matcher.
func NewRouteWithError(name string, priority int, m matcher.Matcher,
	h http.Handler) (Route, error) {
	if name == "" {
		return Route{}, errors.New("the route name is empty")
	}
	if m == nil {
		return Route{}, errors.New("the route matcher is nil")
	}
	if h == nil {
		return Route{}, errors.New("the route handler is nil")
	}
	if priority == 0 {
		priority = m.Priority()
	}
	return Route{Name: name, Priority: priority, Matcher: m, Handler: h}, nil
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
	// BuildMatcherRule is used to build the matcher by the rule string.
	BuildMatcherRule func(matcherRule string) (matcher.Matcher, error)

	notFound handler.SwitchHandler
	handler  handler.SwitchHandler

	glock sync.Mutex
	gmdws handler.Middlewares

	rlock  sync.RWMutex
	rmdws  handler.Middlewares
	origs  map[string]Route
	routes Routes
	router atomic.Value
}

// NewRouter returns a new Router.
func NewRouter() *Router {
	r := &Router{origs: make(map[string]Route, 16)}
	r.handler.Set(http.HandlerFunc(r.serveHTTP))
	r.notFound.Set(handler.Handler404)
	r.router.Store(routesWrapper{})
	return r
}

// SetNotFoundFunc resets the NotFound function as the http handler.
func (r *Router) SetNotFoundFunc(handlerFunc http.HandlerFunc) {
	if handlerFunc == nil {
		panic("the NotFound http handler function is nil")
	}
	r.notFound.Set(handlerFunc)
}

// SetNotFound resets the NotFound handler.
func (r *Router) SetNotFound(handler http.Handler) {
	if handler == nil {
		panic("the NotFound http handler is nil")
	}
	r.notFound.Set(handler)
}

// GetNotFound returns the NotFound handler.
func (r *Router) GetNotFound() http.Handler { return r.notFound.Get() }

// Use appends the http handler middlewars and uses them to act on the route
// handlers when to add the route later.
//
// Notice:
//   - The middlewares will be executed after routing the request.
//   - The middlewares only acts on the routes that will be added later,
//     not the added routes.
//
func (r *Router) Use(mws ...Middleware) {
	r.rlock.Lock()
	r.rmdws.Append(mws...)
	r.rlock.Unlock()
}

// UseReset is the same as Use, but resets the route middlewares to mws.
func (r *Router) UseReset(mws ...Middleware) {
	r.rlock.Lock()
	r.rmdws = append(handler.Middlewares{}, mws...)
	r.rlock.Unlock()
}

// UseCancel removes the http handler middlewares added by Use.
func (r *Router) UseCancel(names ...string) {
	r.rlock.Lock()
	r.rmdws.Remove(names...)
	r.rlock.Unlock()
}

// Global appends the http handler middlewares and uses them to act on
// all the route handlers.
//
// Notice:
//   - The middlewares will be executed before the routing the request.
//   - The middlewares will act on not only the added routes,
//     but also those will be added later.
//
// For example, the log middleware may be used as the global middleware.
func (r *Router) Global(mws ...Middleware) {
	r.glock.Lock()
	defer r.glock.Unlock()
	r.gmdws.Append(mws...)
	r.updateHandler()
}

// GlobalReset is the same as Global, but resets the global middlewares to mws.
func (r *Router) GlobalReset(mws ...Middleware) {
	r.glock.Lock()
	defer r.glock.Unlock()
	r.gmdws = append(handler.Middlewares{}, mws...)
	r.updateHandler()
}

// GlobalCancel removes the http handler middlewares added by Global.
func (r *Router) GlobalCancel(names ...string) {
	r.rlock.Lock()
	defer r.rlock.Unlock()
	r.gmdws.Remove(names...)
	r.updateHandler()
}

func (r *Router) updateHandler() {
	r.handler.Set(r.gmdws.Handler(http.HandlerFunc(r.serveHTTP)))
}

func (r *Router) serveHTTP(w http.ResponseWriter, req *http.Request) {
	routes := r.router.Load().(routesWrapper).Routes
	for i, _len := 0, len(routes); i < _len; i++ {
		if nreq, ok := routes[i].Matcher.Match(req); ok {
			routes[i].handler.ServeHTTP(w, nreq)
			return
		}
	}
	r.notFound.Get().ServeHTTP(w, req)
}

// ServeHTTP implements the interface http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
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
	r.rlock.RLock()
	routes = make(Routes, len(r.routes))
	copy(routes, r.routes)
	r.rlock.RUnlock()
	return
}

// GetRoute returns the route by the name.
func (r *Router) GetRoute(name string) (route Route, ok bool) {
	r.rlock.RLock()
	route, ok = r.origs[name]
	r.rlock.RUnlock()
	return
}

// AddRoute adds the new route.
func (r *Router) AddRoute(route Route) (err error) {
	r.rlock.Lock()
	defer r.rlock.Unlock()
	if _, ok := r.origs[route.Name]; ok {
		err = fmt.Errorf("the route named '%s' has been added", route.Name)
	} else {
		route.Use(r.rmdws...)
		r.origs[route.Name] = route
		r.updateRoutes()
	}
	return
}

// DelRoute deletes and returns the route by the name.
func (r *Router) DelRoute(name string) (route Route, ok bool) {
	r.rlock.Lock()
	if route, ok = r.origs[name]; ok {
		delete(r.origs, name)
		r.updateRoutes()
	}
	r.rlock.Unlock()
	return
}

// DelRoutes deletes all the routes by the names.
func (r *Router) DelRoutes(names ...string) {
	if _len := len(names); _len > 0 {
		r.rlock.Lock()
		for i := 0; i < _len; i++ {
			delete(r.origs, names[i])
		}
		r.updateRoutes()
		r.rlock.Unlock()
	}
}

// UpdateRoutes updates the given routes, which will add the route
// if it does not exist, or update it to the new.
func (r *Router) UpdateRoutes(routes ...Route) {
	if _len := len(routes); _len > 0 {
		r.rlock.Lock()
		defer r.rlock.Unlock()

		for i := 0; i < _len; i++ {
			route := routes[i]
			route.Use(r.rmdws...)
			r.origs[route.Name] = route
		}
		r.updateRoutes()
	}
}

// ResetRoutes discards all the original routes and resets them to the newest.
func (r *Router) ResetRoutes(routes ...Route) {
	_len := len(routes)
	r.rlock.Lock()
	defer r.rlock.Unlock()

	for i := 0; i < _len; i++ {
		route := routes[i]
		route.Use(r.rmdws...)
		r.origs[route.Name] = route
	}
	r.updateRoutes()
}
