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

package ruler

import (
	"errors"
	"net/http"

	"github.com/xgfone/go-apiserver/http/matcher"
	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/middleware"
)

// Name returns a route builder with the name, which is equal to
// NewRouteBuilder(m).Name(name).
func (m *Router) Name(name string) RouteBuilder {
	return NewRouteBuilder(m).Name(name)
}

// Matcher returns a route builder with the matcher,
// which is equal to NewRouteBuilder(m).Matcher(matcher).
func (m *Router) Matcher(matcher matcher.Matcher) RouteBuilder {
	return NewRouteBuilder(m).Matcher(matcher)
}

// Rule returns a route builder with the matcher rule,
// which is equal to NewRouteBuilder(m).Rule(matcherRule).
func (m *Router) Rule(matcherRule string) RouteBuilder {
	return NewRouteBuilder(m).Rule(matcherRule)
}

// Path returns a route builder with the path matcher,
// which is equal to NewRouteBuilder(m).Path(path).
func (m *Router) Path(path string) RouteBuilder {
	return NewRouteBuilder(m).Path(path)
}

// PathPrefix returns a route builder with the path prefix matcher,
// which is equal to NewRouteBuilder(m).PathPrefix(pathPrefix).
func (m *Router) PathPrefix(pathPrefix string) RouteBuilder {
	return NewRouteBuilder(m).PathPrefix(pathPrefix)
}

// Host returns a route builder with the host matcher,
// which is equal to NewRouteBuilder(m).Host(host).
func (m *Router) Host(host string) RouteBuilder {
	return NewRouteBuilder(m).Host(host)
}

// HostRegexp returns a route builder with the host regexp matcher,
// which is equal to NewRouteBuilder(m).HostRegexp(regexpHost).
func (m *Router) HostRegexp(regexpHost string) RouteBuilder {
	return NewRouteBuilder(m).HostRegexp(regexpHost)
}

// RouteBuilder is used to build the route.
type RouteBuilder struct {
	manager  *Router
	mdws     middleware.Middlewares
	name     string
	matcher  matcher.Matcher
	matchers matcher.Matchers
	priority int
	panic    bool
	err      error
}

// NewRouteBuilder returns a new RouteBuilder with the route manager.
func NewRouteBuilder(m *Router) RouteBuilder {
	return RouteBuilder{manager: m, panic: true}
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

// Matcher resets the matcher of the route.
func (b RouteBuilder) Matcher(matcher matcher.Matcher) RouteBuilder {
	b.matcher = matcher
	b.matchers = nil
	b.err = nil
	return b
}

// And appends the matchers based on AND.
func (b RouteBuilder) And(matchers ...matcher.Matcher) RouteBuilder {
	b.matchers = append(b.matchers, matchers...)
	b.matcher = nil
	return b
}

// Or is eqaul to b.And(matcher.Or(matchers...)).
func (b RouteBuilder) Or(matchers ...matcher.Matcher) RouteBuilder {
	return b.And(matcher.Or(matchers...))
}

// Rule is the same as Matcher, but use the builder to build the matcher
// with the matcher rule string.
func (b RouteBuilder) Rule(matcherRule string) RouteBuilder {
	if b.manager.BuildMatcherRule == nil {
		panic("not set the rule buidler of the route matcher")
	}

	if b.err == nil {
		b.matcher, b.err = b.manager.BuildMatcherRule(matcherRule)
		b.matchers = nil
	}
	return b
}

// Path is the same as b.And(matcher.Path(path)).
func (b RouteBuilder) Path(path string) RouteBuilder {
	if b.err == nil {
		var m matcher.Matcher
		if m, b.err = matcher.Path(path); b.err == nil {
			b = b.And(m)
		}
	}
	return b
}

// PathPrefix is the same as b.And(matcher.PathPrefix(pathPrefix)).
func (b RouteBuilder) PathPrefix(pathPrefix string) RouteBuilder {
	if b.err == nil {
		var m matcher.Matcher
		if m, b.err = matcher.PathPrefix(pathPrefix); b.err == nil {
			b = b.And(m)
		}
	}
	return b
}

// Method is the same as b.And(matcher.Method(method)).
func (b RouteBuilder) Method(method string) RouteBuilder {
	if b.err == nil {
		var m matcher.Matcher
		if m, b.err = matcher.Method(method); b.err == nil {
			b = b.And(m)
		}
	}
	return b
}

// ClientIP is the same as b.And(matcher.ClientIP(clientIP)).
func (b RouteBuilder) ClientIP(clientIP string) RouteBuilder {
	if b.err == nil {
		var m matcher.Matcher
		if m, b.err = matcher.ClientIP(clientIP); b.err == nil {
			b = b.And(m)
		}
	}
	return b
}

// Query is the same as b.And(matcher.Query(key, value)).
func (b RouteBuilder) Query(key, value string) RouteBuilder {
	if b.err == nil {
		var m matcher.Matcher
		if m, b.err = matcher.Query(key, value); b.err == nil {
			b = b.And(m)
		}
	}
	return b
}

// Header is the same as b.And(matcher.Header(key, value)).
func (b RouteBuilder) Header(key, value string) RouteBuilder {
	if b.err == nil {
		var m matcher.Matcher
		if m, b.err = matcher.Header(key, value); b.err == nil {
			b = b.And(m)
		}
	}
	return b
}

// HeaderRegexp is the same as b.And(matcher.HeaderRegexp(key, regexpValue)).
func (b RouteBuilder) HeaderRegexp(key, regexpValue string) RouteBuilder {
	if b.err == nil {
		var m matcher.Matcher
		if m, b.err = matcher.HeaderRegexp(key, regexpValue); b.err == nil {
			b = b.And(m)
		}
	}
	return b
}

// Host is the same as b.And(matcher.Host(host)).
func (b RouteBuilder) Host(host string) RouteBuilder {
	if b.err == nil {
		var m matcher.Matcher
		if m, b.err = matcher.Host(host); b.err == nil {
			b = b.And(m)
		}
	}
	return b
}

// HostRegexp is the same as b.And(matcher.HostRegexp(regexpHost)).
func (b RouteBuilder) HostRegexp(regexpHost string) RouteBuilder {
	if b.err == nil {
		var m matcher.Matcher
		if m, b.err = matcher.HostRegexp(regexpHost); b.err == nil {
			b = b.And(m)
		}
	}
	return b
}

// Use appends the http handler middlewares that act on the latter handler.
func (b RouteBuilder) Use(middlewares ...middleware.Middleware) RouteBuilder {
	b.mdws = b.mdws.Clone(middlewares...)
	return b
}

// ContextHandler is the same HandlerFunc, but wraps the request and response
// into Context.
func (b RouteBuilder) ContextHandler(h reqresp.Handler) error {
	return b.Handler(h)
}

// ContextHandlerWithError is the same ContextHandler, but supports to return
// an error.
func (b RouteBuilder) ContextHandlerWithError(h reqresp.HandlerWithError) error {
	return b.Handler(h)
}

// HandlerFunc registers the route with the handler functions.
func (b RouteBuilder) HandlerFunc(handler http.HandlerFunc) error {
	return b.Handler(handler)
}

// Handler registers the route with the handler.
func (b RouteBuilder) Handler(handler http.Handler) (err error) {
	if err = b.err; err == nil {
		err = b.addRoute(handler)
	}

	if err != nil && b.panic {
		panic(err)
	}
	return
}

func (b RouteBuilder) addRoute(handler http.Handler) (err error) {
	if b.matcher == nil && len(b.matchers) > 0 {
		b.matcher = matcher.And(b.matchers...)
	}

	if b.matcher == nil {
		return errors.New("mising the route matcher")
	}

	name := b.name
	if name == "" {
		name = b.matcher.String()
	}

	handler = b.mdws.Handler(handler).(http.Handler)
	route := NewRoute(name, b.priority, b.matcher, handler)
	return b.manager.AddRoute(route)
}

// GET is a convenient function to register the route with the handler,
// which is the same as b.Method(http.MethodGet).Handler(handler).
func (b RouteBuilder) GET(handler http.Handler) RouteBuilder {
	b.Method(http.MethodGet).SetPanic(true).Handler(handler)
	return b
}

// PUT is a convenient function to register the route with the handler,
// which is the same as b.Method(http.MethodPut).Handler(handler).
func (b RouteBuilder) PUT(handler http.Handler) RouteBuilder {
	b.Method(http.MethodPut).SetPanic(true).Handler(handler)
	return b
}

// POST is a convenient function to register the route with the handler,
// which is the same as b.Method(http.MethodPost).Handler(handler).
func (b RouteBuilder) POST(handler http.Handler) RouteBuilder {
	b.Method(http.MethodPost).SetPanic(true).Handler(handler)
	return b
}

// DELETE is a convenient function to register the route with the handler,
// which is the same as b.Method(http.MethodDelete).Handler(handler).
func (b RouteBuilder) DELETE(handler http.Handler) RouteBuilder {
	b.Method(http.MethodDelete).SetPanic(true).Handler(handler)
	return b
}

// PATCH is a convenient function to register the route with the handler,
// which is the same as b.Method(http.MethodPatch).Handler(handler).
func (b RouteBuilder) PATCH(handler http.Handler) RouteBuilder {
	b.Method(http.MethodPatch).SetPanic(true).Handler(handler)
	return b
}

// HEAD is a convenient function to register the route with the handler,
// which is the same as b.Method(http.MethodHead).Handler(handler).
func (b RouteBuilder) HEAD(handler http.Handler) RouteBuilder {
	b.Method(http.MethodHead).SetPanic(true).Handler(handler)
	return b
}

// GETFunc is a convenient function to register the route with the function
// handler, which is the same as b.Method(http.MethodGet).Handler(handler).
func (b RouteBuilder) GETFunc(handler http.HandlerFunc) RouteBuilder {
	b.Method(http.MethodGet).SetPanic(true).Handler(handler)
	return b
}

// PUTFunc is a convenient function to register the route with the function
// handler, which is the same as b.Method(http.MethodPut).Handler(handler).
func (b RouteBuilder) PUTFunc(handler http.HandlerFunc) RouteBuilder {
	b.Method(http.MethodPut).SetPanic(true).Handler(handler)
	return b
}

// POSTFunc is a convenient function to register the route with the function
// handler, which is the same as b.Method(http.MethodPost).Handler(handler).
func (b RouteBuilder) POSTFunc(handler http.HandlerFunc) RouteBuilder {
	b.Method(http.MethodPost).SetPanic(true).Handler(handler)
	return b
}

// DELETEFunc is a convenient function to register the route with the function
// handler, which is the same as b.Method(http.MethodDelete).Handler(handler).
func (b RouteBuilder) DELETEFunc(handler http.HandlerFunc) RouteBuilder {
	b.Method(http.MethodDelete).SetPanic(true).Handler(handler)
	return b
}

// PATCHFunc is a convenient function to register the route with the function
// handler, which is the same as b.Method(http.MethodPatch).Handler(handler).
func (b RouteBuilder) PATCHFunc(handler http.HandlerFunc) RouteBuilder {
	b.Method(http.MethodPatch).SetPanic(true).Handler(handler)
	return b
}

// HEADFunc is a convenient function to register the route with the function
// handler, which is the same as b.Method(http.MethodHead).Handler(handler).
func (b RouteBuilder) HEADFunc(handler http.HandlerFunc) RouteBuilder {
	b.Method(http.MethodHead).SetPanic(true).Handler(handler)
	return b
}

// GETContext is a convenient function to register the route with the context
// handler, which is the same as b.Method(http.MethodGet).Handler(handler).
func (b RouteBuilder) GETContext(handler reqresp.Handler) RouteBuilder {
	b.Method(http.MethodGet).SetPanic(true).Handler(handler)
	return b
}

// PUTContext is a convenient function to register the route with the context
// handler, which is the same as b.Method(http.MethodPut).Handler(handler).
func (b RouteBuilder) PUTContext(handler reqresp.Handler) RouteBuilder {
	b.Method(http.MethodPut).SetPanic(true).Handler(handler)
	return b
}

// POSTContext is a convenient function to register the route with the context
// handler, which is the same as b.Method(http.MethodPost).Handler(handler).
func (b RouteBuilder) POSTContext(handler reqresp.Handler) RouteBuilder {
	b.Method(http.MethodPost).SetPanic(true).Handler(handler)
	return b
}

// DELETEContext is a convenient function to register the route with the context
// handler, which is the same as b.Method(http.MethodDelete).Handler(handler).
func (b RouteBuilder) DELETEContext(handler reqresp.Handler) RouteBuilder {
	b.Method(http.MethodDelete).SetPanic(true).Handler(handler)
	return b
}

// PATCHContext is a convenient function to register the route with the context
// handler, which is the same as b.Method(http.MethodPatch).Handler(handler).
func (b RouteBuilder) PATCHContext(handler reqresp.Handler) RouteBuilder {
	b.Method(http.MethodPatch).SetPanic(true).Handler(handler)
	return b
}

// HEADContext is a convenient function to register the route with the context
// handler, which is the same as b.Method(http.MethodHead).Handler(handler).
func (b RouteBuilder) HEADContext(handler reqresp.Handler) RouteBuilder {
	b.Method(http.MethodHead).SetPanic(true).Handler(handler)
	return b
}

// GETContextWithError is a convenient function to register the route with the
// context handler, which is the same as b.Method(http.MethodGet).Handler(handler).
func (b RouteBuilder) GETContextWithError(handler reqresp.HandlerWithError) RouteBuilder {
	b.Method(http.MethodGet).SetPanic(true).Handler(handler)
	return b
}

// PUTContextWithError is a convenient function to register the route with the
// context handler, which is the same as b.Method(http.MethodPut).Handler(handler).
func (b RouteBuilder) PUTContextWithError(handler reqresp.HandlerWithError) RouteBuilder {
	b.Method(http.MethodPut).SetPanic(true).Handler(handler)
	return b
}

// POSTContextWithError is a convenient function to register the route with the
// context handler, which is the same as b.Method(http.MethodPost).Handler(handler).
func (b RouteBuilder) POSTContextWithError(handler reqresp.HandlerWithError) RouteBuilder {
	b.Method(http.MethodPost).SetPanic(true).Handler(handler)
	return b
}

// DELETEContextWithError is a convenient function to register the route with the
// context handler, which is the same as b.Method(http.MethodDelete).Handler(handler).
func (b RouteBuilder) DELETEContextWithError(handler reqresp.HandlerWithError) RouteBuilder {
	b.Method(http.MethodDelete).SetPanic(true).Handler(handler)
	return b
}

// PATCHContextWithError is a convenient function to register the route with the
// context handler, which is the same as b.Method(http.MethodPatch).Handler(handler).
func (b RouteBuilder) PATCHContextWithError(handler reqresp.HandlerWithError) RouteBuilder {
	b.Method(http.MethodPatch).SetPanic(true).Handler(handler)
	return b
}

// HEADContextWithError is a convenient function to register the route with the
// context handler, which is the same as b.Method(http.MethodHead).Handler(handler).
func (b RouteBuilder) HEADContextWithError(handler reqresp.HandlerWithError) RouteBuilder {
	b.Method(http.MethodHead).SetPanic(true).Handler(handler)
	return b
}
