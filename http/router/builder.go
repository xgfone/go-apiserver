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

package router

import (
	"fmt"
	"net/http"

	"github.com/xgfone/go-apiserver/http/matcher"
)

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

func (b RouteBuilder) addRoute(handler http.Handler) (err error) {
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

	if b.matcher == nil {
		if b.matcher, err = b.router.builder.Parse(rule); err != nil {
			return err
		}
	}

	route, err := NewRouteWithError(name, b.priority, b.matcher, handler)
	if err != nil {
		return err
	}

	return b.router.AddRoute(route)
}
