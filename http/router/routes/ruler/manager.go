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
	"fmt"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/xgfone/go-apiserver/http/handler"
	"github.com/xgfone/go-apiserver/http/matcher"
	"github.com/xgfone/go-apiserver/internal/ruler"
)

type routesWrapper struct{ Routes }

// RouteManager is used to manage a set of routes based on the ruler.
type RouteManager struct {
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

// NewRouteManager returns a new route manager.
func NewRouteManager() *RouteManager {
	m := &RouteManager{rmaps: make(map[string]Route, 16)}
	m.BuildMatcherRule = ruler.Build
	m.updateRoutes()
	return m
}

// ServeHTTP implements the interface http.Handler.
func (m *RouteManager) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	m.Route(resp, req, m.NotFound)
}

// Route implements the interface router.RouteManager.
func (m *RouteManager) Route(w http.ResponseWriter, r *http.Request, notFound http.Handler) {
	if route, ok := m.MatchRoute(w, r); ok {
		route.ServeHTTP(w, r)
	} else if notFound != nil {
		notFound.ServeHTTP(w, r)
	} else if m.NotFound != nil {
		m.NotFound.ServeHTTP(w, r)
	} else {
		handler.Handler404.ServeHTTP(w, r)
	}
}

/* ------------------------------------------------------------------------ */

// MatchRoute uses the registered routes to match the http request,
// and returns the matched route.
func (m *RouteManager) MatchRoute(w http.ResponseWriter, r *http.Request) (Route, bool) {
	routes := m.routes.Load().(routesWrapper).Routes
	for i, _len := 0, len(routes); i < _len; i++ {
		if ok := routes[i].Matcher.Match(w, r); ok {
			return routes[i], true
		}
	}
	return Route{}, false
}

// GetRoute returns the route by the given name.
func (m *RouteManager) GetRoute(name string) (route Route, ok bool) {
	m.rlock.RLock()
	route, ok = m.rmaps[name]
	m.rlock.RUnlock()
	return
}

// GetRoutes returns all the registered routes.
func (m *RouteManager) GetRoutes() (routes Routes) {
	origs := m.routes.Load().(routesWrapper).Routes
	routes = make(Routes, len(origs))
	copy(routes, origs)
	return
}

// AddRoute adds the given route.
func (m *RouteManager) AddRoute(route Route) (err error) {
	if err = m.checkRoute(route); err != nil {
		return
	}

	m.rlock.Lock()
	if _, ok := m.rmaps[route.Name]; ok {
		err = fmt.Errorf("the route named '%s' has been added", route.Name)
	} else {
		m.rmaps[route.Name] = route
		m.updateRoutes()
	}
	m.rlock.Unlock()
	return
}

// DelRoute deletes and returns the route by the given name.
func (m *RouteManager) DelRoute(name string) (route Route, ok bool) {
	if len(name) == 0 {
		return
	}

	m.rlock.Lock()
	if route, ok = m.rmaps[name]; ok {
		delete(m.rmaps, name)
		m.updateRoutes()
	}
	m.rlock.Unlock()
	return
}

// UpdateRoutes updates the given routes, which will add the route
// if it does not exist, or update it to the new.
func (m *RouteManager) UpdateRoutes(routes ...Route) (err error) {
	if err = m.checkRoutes(routes); err != nil {
		return
	}

	if _len := len(routes); _len > 0 {
		m.rlock.Lock()
		for i := 0; i < _len; i++ {
			route := routes[i]
			m.rmaps[route.Name] = route
		}
		m.updateRoutes()
		m.rlock.Unlock()
	}
	return
}

// ResetRoutes discards all the original routes and resets them to routes.
func (m *RouteManager) ResetRoutes(routes ...Route) (err error) {
	if err = m.checkRoutes(routes); err != nil {
		return
	}

	m.rlock.Lock()
	for name := range m.rmaps {
		delete(m.rmaps, name)
	}

	for i, _len := 0, len(routes); i < _len; i++ {
		route := routes[i]
		m.rmaps[route.Name] = route
	}
	m.updateRoutes()
	m.rlock.Unlock()
	return
}

func (m *RouteManager) checkRoutes(routes Routes) (err error) {
	for _len := len(routes) - 1; _len >= 0; _len-- {
		if err = m.checkRoute(routes[_len]); err != nil {
			break
		}
	}
	return
}

func (m *RouteManager) checkRoute(route Route) error {
	if len(route.Name) == 0 {
		return errors.New("the route name is empty")
	} else if route.Handler == nil {
		return errors.New("the route handler is nil")
	} else if route.Matcher == nil {
		return errors.New("the route matcher is nil")
	}
	return nil
}

func (m *RouteManager) updateRoutes() {
	routes := make(Routes, 0, len(m.rmaps))
	for _, route := range m.rmaps {
		routes = append(routes, route)
	}
	sort.Stable(routes)
	m.routes.Store(routesWrapper{Routes: routes})
}
