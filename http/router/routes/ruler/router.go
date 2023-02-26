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

// Package ruler implements a route manager based on the ruler.
package ruler

import (
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"
	rpprof "runtime/pprof"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/xgfone/go-apiserver/http/handler"
	"github.com/xgfone/go-apiserver/http/header"
	"github.com/xgfone/go-apiserver/http/matcher"
	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/http/router/routes/action"
	"github.com/xgfone/go-apiserver/internal/ruler"
	"github.com/xgfone/go-apiserver/tools/maps"
	"github.com/xgfone/go-apiserver/tools/slices"
)

// DefaultRouter is the default global ruler router.
var DefaultRouter = NewRouter()

type routesWrapper struct{ Routes }

// Router is used to manage a set of routes based on the ruler.
type Router struct {
	// BuildMatcherRule is used to build the matcher by the rule string.
	//
	// Default: ruler.Build
	// See https://pkg.go.dev/github.com/xgfone/go-apiserver/internal/ruler#Build
	BuildMatcherRule func(matcherRule string) (matcher.Matcher, error)

	// NotFound is used when the manager is used as http.Handler
	// and does not find the route.
	//
	// Default: handler.Handler404
	NotFound http.Handler

	rlock  sync.RWMutex
	rmaps  map[string]Route
	routes atomic.Value
}

// NewRouter returns a new route manager.
func NewRouter() *Router {
	r := &Router{rmaps: make(map[string]Route, 16)}
	r.BuildMatcherRule = ruler.Build
	r.updateRoutes()
	return r
}

// ServeHTTP implements the interface http.Handler.
func (r *Router) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	r.Route(resp, req, r.NotFound)
}

// Route is the same ServeHTTP, but support to provide a NotFound handler.
func (r *Router) Route(rw http.ResponseWriter, req *http.Request, notFound http.Handler) {
	if route, ok := r.MatchRoute(rw, req); ok {
		route.ServeHTTP(rw, req)
	} else if notFound != nil {
		notFound.ServeHTTP(rw, req)
	} else if r.NotFound != nil {
		r.NotFound.ServeHTTP(rw, req)
	} else {
		handler.Handler404.ServeHTTP(rw, req)
	}
}

/* ------------------------------------------------------------------------ */

// MatchRoute uses the registered routes to match the http request,
// and returns the matched route.
func (r *Router) MatchRoute(rw http.ResponseWriter, req *http.Request) (Route, bool) {
	routes := r.routes.Load().(routesWrapper).Routes
	for i, _len := 0, len(routes); i < _len; i++ {
		if ok := routes[i].Matcher.Match(rw, req); ok {
			return routes[i], true
		}
	}
	return Route{}, false
}

// GetRoute returns the route by the given name.
func (r *Router) GetRoute(name string) (route Route, ok bool) {
	r.rlock.RLock()
	route, ok = r.rmaps[name]
	r.rlock.RUnlock()
	return
}

// GetRoutes returns all the registered routes.
func (r *Router) GetRoutes() (routes Routes) {
	return slices.Clone(r.routes.Load().(routesWrapper).Routes)
}

// AddRoute adds the given route.
func (r *Router) AddRoute(route Route) (err error) {
	if err = r.checkRoute(route); err != nil {
		return
	}

	r.rlock.Lock()
	if maps.Add(r.rmaps, route.Name, route) {
		r.updateRoutes()
	} else {
		err = fmt.Errorf("the route named '%s' has been added", route.Name)
	}
	r.rlock.Unlock()

	return
}

// DelRoute deletes and returns the route by the given name.
func (r *Router) DelRoute(name string) (route Route, ok bool) {
	if len(name) == 0 {
		return
	}

	r.rlock.Lock()
	if route, ok = maps.Pop(r.rmaps, name); ok {
		r.updateRoutes()
	}
	r.rlock.Unlock()
	return
}

// UpdateRoutes updates the given routes, which will add the route
// if it does not exist, or update it to the new.
func (r *Router) UpdateRoutes(routes ...Route) (err error) {
	if err = r.checkRoutes(routes); err != nil {
		return
	}

	if _len := len(routes); _len > 0 {
		r.rlock.Lock()
		maps.AddSliceAsValue(r.rmaps, routes, func(r Route) string { return r.Name })
		r.updateRoutes()
		r.rlock.Unlock()
	}
	return
}

// ResetRoutes discards all the original routes and resets them to routes.
func (r *Router) ResetRoutes(routes ...Route) (err error) {
	if err = r.checkRoutes(routes); err != nil {
		return
	}

	r.rlock.Lock()
	maps.Clear(r.rmaps)
	maps.AddSliceAsValue(r.rmaps, routes, func(r Route) string { return r.Name })
	r.updateRoutes()
	r.rlock.Unlock()
	return
}

func (r *Router) checkRoutes(routes Routes) (err error) {
	for _len := len(routes) - 1; _len >= 0; _len-- {
		if err = r.checkRoute(routes[_len]); err != nil {
			break
		}
	}
	return
}

func (r *Router) checkRoute(route Route) error {
	if len(route.Name) == 0 {
		return errors.New("the route name is empty")
	} else if route.Handler == nil {
		return errors.New("the route handler is nil")
	} else if route.Matcher == nil {
		return errors.New("the route matcher is nil")
	}
	return nil
}

func (r *Router) updateRoutes() {
	routes := Routes(maps.Values(r.rmaps))
	sort.Stable(routes)
	r.routes.Store(routesWrapper{Routes: routes})
}

// AddVarsRoute adds the route to serve the published vars coming from "expvar".
func (r *Router) AddVarsRoute(pathPrefix string) {
	pathPrefix = strings.TrimRight(pathPrefix, "/")
	r.Path(pathPrefix + "/debug/vars").GET(expvar.Handler())
}

// AddProfileRoutes adds the profile routes coming from "net/http/pprof".
func (r *Router) AddProfileRoutes(pathPrefix string) {
	pathPrefix = strings.TrimRight(pathPrefix, "/")
	r.Path(pathPrefix + "/debug/pprof/profile").GETFunc(pprof.Profile)
	r.Path(pathPrefix + "/debug/pprof/cmdline").GETFunc(pprof.Cmdline)
	r.Path(pathPrefix + "/debug/pprof/symbol").GETFunc(pprof.Symbol)
	r.Path(pathPrefix + "/debug/pprof/trace").GETFunc(pprof.Trace)
	r.Path(pathPrefix + "/debug/pprof/").GETFunc(pprof.Index)
	r.Path(pathPrefix + "/debug/pprof").GETFunc(pprofIndex(pathPrefix))
	for _, p := range rpprof.Profiles() {
		r.Path(pathPrefix + "/debug/pprof/" + p.Name()).GET(pprof.Handler(p.Name()))
	}
}

func pprofIndex(prefix string) func(w http.ResponseWriter, r *http.Request) {
	path := prefix + "/debug/pprof/"
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(header.HeaderLocation, path)
		w.WriteHeader(301)
	}
}

// AddRuleRoute adds a route to serve the routes of the current router
// with the path pathPrefix+"/debug/router/ruler/routes".
func (r *Router) AddRuleRoute(pathPrefix string) {
	pathPrefix = strings.TrimRight(pathPrefix, "/")
	r.Path(pathPrefix + "/debug/router/ruler/routes").GETContextWithError(func(c *reqresp.Context) error {
		return c.JSON(200, map[string]interface{}{"Routes": r.GetRoutes()})
	})
}

// AddActionRoute adds a route to serve the actions of the action router
// with the path pathPrefix+"/debug/router/action/actions".
//
// If the action router is nil, use action.DefaultRouter instead.
func (r *Router) AddActionRoute(pathPrefix string, router *action.Router) {
	pathPrefix = strings.TrimRight(pathPrefix, "/")
	r.Path(pathPrefix + "/debug/router/action/actions").GETContext(func(c *reqresp.Context) {
		r := router
		if r == nil {
			r = action.DefaultRouter
		}

		c.JSON(200, map[string][]string{"Actions": r.GetActions()})
	})
}
