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

package ruler

import (
	"net/http"
	"slices"
	"strings"

	"github.com/xgfone/go-apiserver/http/middleware"
	"github.com/xgfone/go-apiserver/http/reqresp"
	matcher "github.com/xgfone/go-http-matcher"
)

// Group returns a route builder with the prefix path group,
// which will register the built route into the router.
func (r *Router) Group(pathPrefix string) RouteBuilder {
	return r.RouteBuilder().Group(pathPrefix)
}

// Path returns a route builder with the path matcher,
// which will register the built route into the router.
func (r *Router) Path(path string) RouteBuilder {
	return r.RouteBuilder().Path(path)
}

// PathPrefix returns a route builder with the path prefix matcher,
// which will register the built route into the router.
func (r *Router) PathPrefix(pathPrefix string) RouteBuilder {
	return r.RouteBuilder().PathPrefix(pathPrefix)
}

// Host returns a route builder with the host matcher,
// which will register the built route into the router.
func (r *Router) Host(host string) RouteBuilder {
	return r.RouteBuilder().Host(host)
}

// Route returns a new route builder.
func (r *Router) RouteBuilder() RouteBuilder {
	return NewRouteBuilder(r.Register)
}

// RouteBuilder is used to build the route.
type RouteBuilder struct {
	register func(Route)

	mdws  middleware.Middlewares
	group string
	route Route

	host    matcher.Matcher
	path    matcher.Matcher
	method  matcher.Matcher
	headers []matcher.Matcher
	queries []matcher.Matcher
}

// NewRouteBuilder returns a new route builder.
func NewRouteBuilder(register func(Route)) RouteBuilder {
	return RouteBuilder{register: register}
}

func appendMatcher(ms []matcher.Matcher, m matcher.Matcher) []matcher.Matcher {
	if m != nil {
		matchers := make([]matcher.Matcher, 0, len(ms)+1)
		matchers = append(matchers, ms...)
		matchers = append(matchers, m)
		ms = matchers
	}
	return ms
}

// Prefix returns the group path prefix, which is not the path prefix matcher.
func (b RouteBuilder) Prefix() string { return b.group }

// Use appends the http handler middlewares that act on the later handler.
func (b RouteBuilder) Use(middlewares ...middleware.Middleware) RouteBuilder {
	b.mdws = b.mdws.Append(middlewares...)
	return b
}

// UseFunc appends the http handler function middlewares that act on the later handler.
func (b RouteBuilder) UseFunc(middlewares ...middleware.MiddlewareFunc) RouteBuilder {
	b.mdws = b.mdws.AppendFunc(middlewares...)
	return b
}

// Clone clones itself and returns a new route builder.
func (b RouteBuilder) Clone() RouteBuilder {
	b.headers = slices.Clone(b.headers)
	b.queries = slices.Clone(b.queries)
	b.mdws = b.mdws.Clone()
	return b
}

// Extra sets the extra data of the route.
func (b RouteBuilder) Extra(extra interface{}) RouteBuilder {
	b.route.Extra = extra
	return b
}

// Group appends the prefix of the paths of a group of routes
// when they are registered,
func (b RouteBuilder) Group(pathPrefix string) RouteBuilder {
	pathPrefix = strings.TrimRight(pathPrefix, "/")
	if pathPrefix != "" && pathPrefix[0] != '/' {
		pathPrefix = "/" + pathPrefix
	}

	if b.group == "" {
		b.group = pathPrefix
	} else if pathPrefix != "" {
		b.group += pathPrefix
	}
	return b
}

// Priority sets the priority the route.
func (b RouteBuilder) Priority(priority int) RouteBuilder {
	b.route.Priority = priority
	return b
}

// Desc sets the description of the route.
func (b RouteBuilder) Desc(desc string) RouteBuilder {
	b.route.Desc = desc
	return b
}

/// ----------------------------------------------------------------------- ///
// Matcher

// Path adds the path match rule, which ignores the trailling "/".
//
//   - If the group is set, it will add it into path as the prefix.
//   - It supports the path parameters, such as "/prefix/{param1}/path/{param2}/to",
//     and put the parsed parameter values into the Data field
//     if a *reqresp.Context can be got from *http.Request.
func (b RouteBuilder) Path(path string) RouteBuilder {
	if path == "" {
		return b
	}

	if path[0] != '/' {
		path = "/" + path
	}

	if b.group != "" {
		if path == "/" {
			path = b.group
		} else {
			path = b.group + path
		}
	}

	b.path = newPathMatcher(path)
	return b
}

// PathPrefix adds the path prefeix match rule, which ignores the trailling "/".
//
//   - If the group is set, it will add it into pathPrefix as the prefix.
//   - It supports the path parameters, such as "/prefix/{param1}/path/{param2}/to",
//     and put the parsed parameter values into the Data field
//     if a *reqresp.Context can be got from *http.Request.
func (b RouteBuilder) PathPrefix(pathPrefix string) RouteBuilder {
	if pathPrefix == "" {
		return b
	}

	if pathPrefix[0] != '/' {
		pathPrefix = "/" + pathPrefix
	}

	if b.group != "" {
		if pathPrefix == "/" {
			pathPrefix = b.group
		} else {
			pathPrefix = b.group + pathPrefix
		}
	}

	b.path = newPathPrefixMatcher(pathPrefix)
	return b
}

// Method adds the method match ruler.
func (b RouteBuilder) Method(method string) RouteBuilder {
	b.method = matcher.Method(method)
	return b
}

// Query adds the query key-value match ruler.
func (b RouteBuilder) Query(key, value string) RouteBuilder {
	b.queries = appendMatcher(b.queries, matcher.Query(key, value))
	return b
}

// Header adds the header key-value match ruler.
func (b RouteBuilder) Header(key, value string) RouteBuilder {
	b.headers = appendMatcher(b.headers, matcher.Header(key, value))
	return b
}

// Host adds the host match ruler.
func (b RouteBuilder) Host(host string) RouteBuilder {
	b.host = matcher.Host(host)
	return b
}

/// ----------------------------------------------------------------------- ///
// Handler & Register

// Handler registers the route with the handler.
func (b RouteBuilder) Handler(handler http.Handler) RouteBuilder {
	b.register(b.newRoute(handler))
	return b
}

func tryAppendMatcher(ms []matcher.Matcher, m matcher.Matcher) []matcher.Matcher {
	if m != nil {
		ms = append(ms, m)
	}
	return ms
}

func (b RouteBuilder) newRoute(handler http.Handler) (route Route) {
	var headers, queries matcher.Matcher
	if len(b.headers) > 0 {
		headers = matcher.And(b.headers...)
	}
	if len(b.queries) > 0 {
		queries = matcher.And(b.queries...)
	}

	matchers := make([]matcher.Matcher, 0, 4)
	matchers = tryAppendMatcher(matchers, b.host)
	matchers = tryAppendMatcher(matchers, b.path)
	matchers = tryAppendMatcher(matchers, b.method)
	matchers = tryAppendMatcher(matchers, headers)
	matchers = tryAppendMatcher(matchers, queries)
	matcher := matcher.And(matchers...)

	route = b.route
	route.Matcher = matcher
	route.Handler = handler
	route.Use(b.mdws)

	if route.Priority == 0 {
		route.Priority = matcher.Priority()
	}
	if route.Desc == "" {
		route.Desc = matcher.String()
	}

	return
}

/// ----------------------------------------------------------------------- ///
// For http.Handler

// GET is a convenient function to register the route with the handler,
// which is the same as b.Method(http.MethodGet).Handler(handler).
func (b RouteBuilder) GET(handler http.Handler) RouteBuilder {
	return b.Method(http.MethodGet).Handler(handler)
}

// PUT is a convenient function to register the route with the handler,
// which is the same as b.Method(http.MethodPut).Handler(handler).
func (b RouteBuilder) PUT(handler http.Handler) RouteBuilder {
	return b.Method(http.MethodPut).Handler(handler)
}

// POST is a convenient function to register the route with the handler,
// which is the same as b.Method(http.MethodPost).Handler(handler).
func (b RouteBuilder) POST(handler http.Handler) RouteBuilder {
	return b.Method(http.MethodPost).Handler(handler)
}

// DELETE is a convenient function to register the route with the handler,
// which is the same as b.Method(http.MethodDelete).Handler(handler).
func (b RouteBuilder) DELETE(handler http.Handler) RouteBuilder {
	return b.Method(http.MethodDelete).Handler(handler)
}

// PATCH is a convenient function to register the route with the handler,
// which is the same as b.Method(http.MethodPatch).Handler(handler).
func (b RouteBuilder) PATCH(handler http.Handler) RouteBuilder {
	return b.Method(http.MethodPatch).Handler(handler)
}

// HEAD is a convenient function to register the route with the handler,
// which is the same as b.Method(http.MethodHead).Handler(handler).
func (b RouteBuilder) HEAD(handler http.Handler) RouteBuilder {
	return b.Method(http.MethodHead).Handler(handler)
}

// OPTIONS is a convenient function to register the route with the handler,
// which is the same as b.Method(http.MethodOptions).Handler(handler).
func (b RouteBuilder) OPTIONS(handler http.Handler) RouteBuilder {
	return b.Method(http.MethodOptions).Handler(handler)
}

/// ----------------------------------------------------------------------- ///
// For http.HandlerFunc

// HandlerFunc registers the route with the handler functions.
func (b RouteBuilder) HandlerFunc(handler http.HandlerFunc) RouteBuilder {
	return b.Handler(handler)
}

// GETFunc is a convenient function to register the route with the function
// handler, which is the same as b.Method(http.MethodGet).Handler(handler).
func (b RouteBuilder) GETFunc(handler http.HandlerFunc) RouteBuilder {
	b.Method(http.MethodGet).Handler(handler)
	return b
}

// PUTFunc is a convenient function to register the route with the function
// handler, which is the same as b.Method(http.MethodPut).Handler(handler).
func (b RouteBuilder) PUTFunc(handler http.HandlerFunc) RouteBuilder {
	b.Method(http.MethodPut).Handler(handler)
	return b
}

// POSTFunc is a convenient function to register the route with the function
// handler, which is the same as b.Method(http.MethodPost).Handler(handler).
func (b RouteBuilder) POSTFunc(handler http.HandlerFunc) RouteBuilder {
	b.Method(http.MethodPost).Handler(handler)
	return b
}

// DELETEFunc is a convenient function to register the route with the function
// handler, which is the same as b.Method(http.MethodDelete).Handler(handler).
func (b RouteBuilder) DELETEFunc(handler http.HandlerFunc) RouteBuilder {
	b.Method(http.MethodDelete).Handler(handler)
	return b
}

// PATCHFunc is a convenient function to register the route with the function
// handler, which is the same as b.Method(http.MethodPatch).Handler(handler).
func (b RouteBuilder) PATCHFunc(handler http.HandlerFunc) RouteBuilder {
	b.Method(http.MethodPatch).Handler(handler)
	return b
}

// HEADFunc is a convenient function to register the route with the function
// handler, which is the same as b.Method(http.MethodHead).Handler(handler).
func (b RouteBuilder) HEADFunc(handler http.HandlerFunc) RouteBuilder {
	b.Method(http.MethodHead).Handler(handler)
	return b
}

// OPTIONS is a convenient function to register the route with the handler,
// which is the same as b.Method(http.MethodOptions).Handler(handler).
func (b RouteBuilder) OPTIONSFunc(handler http.HandlerFunc) RouteBuilder {
	return b.Method(http.MethodOptions).Handler(handler)
}

/// ----------------------------------------------------------------------- ///
// For Context

// ContextHandler is the same HandlerFunc, but wraps the request and response into Context.
func (b RouteBuilder) ContextHandler(h reqresp.Handler) RouteBuilder {
	return b.Handler(h)
}

// GETContext is a convenient function to register the route with the context
// handler, which is the same as b.Method(http.MethodGet).Handler(handler).
func (b RouteBuilder) GETContext(handler reqresp.Handler) RouteBuilder {
	return b.Method(http.MethodGet).Handler(handler)
}

// PUTContext is a convenient function to register the route with the context
// handler, which is the same as b.Method(http.MethodPut).Handler(handler).
func (b RouteBuilder) PUTContext(handler reqresp.Handler) RouteBuilder {
	return b.Method(http.MethodPut).Handler(handler)
}

// POSTContext is a convenient function to register the route with the context
// handler, which is the same as b.Method(http.MethodPost).Handler(handler).
func (b RouteBuilder) POSTContext(handler reqresp.Handler) RouteBuilder {
	return b.Method(http.MethodPost).Handler(handler)
}

// DELETEContext is a convenient function to register the route with the context
// handler, which is the same as b.Method(http.MethodDelete).Handler(handler).
func (b RouteBuilder) DELETEContext(handler reqresp.Handler) RouteBuilder {
	return b.Method(http.MethodDelete).Handler(handler)
}

// PATCHContext is a convenient function to register the route with the context
// handler, which is the same as b.Method(http.MethodPatch).Handler(handler).
func (b RouteBuilder) PATCHContext(handler reqresp.Handler) RouteBuilder {
	return b.Method(http.MethodPatch).Handler(handler)
}

// HEADContext is a convenient function to register the route with the context
// handler, which is the same as b.Method(http.MethodHead).Handler(handler).
func (b RouteBuilder) HEADContext(handler reqresp.Handler) RouteBuilder {
	return b.Method(http.MethodHead).Handler(handler)
}

// OPTIONSContext is a convenient function to register the route with the context
// handler, which is the same as b.Method(http.MethodOptions).Handler(handler).
func (b RouteBuilder) OPTIONSContext(handler reqresp.Handler) RouteBuilder {
	return b.Method(http.MethodOptions).Handler(handler)
}

/// ----------------------------------------------------------------------- ///
// For ContextWithError

// ContextHandlerWithError is the same ContextHandler, but supports to return an error.
func (b RouteBuilder) ContextHandlerWithError(h reqresp.HandlerWithError) RouteBuilder {
	return b.Handler(h)
}

// GETContextWithError is a convenient function to register the route with the
// context handler, which is the same as b.Method(http.MethodGet).Handler(handler).
func (b RouteBuilder) GETContextWithError(handler reqresp.HandlerWithError) RouteBuilder {
	return b.Method(http.MethodGet).Handler(handler)
}

// PUTContextWithError is a convenient function to register the route with the
// context handler, which is the same as b.Method(http.MethodPut).Handler(handler).
func (b RouteBuilder) PUTContextWithError(handler reqresp.HandlerWithError) RouteBuilder {
	return b.Method(http.MethodPut).Handler(handler)
}

// POSTContextWithError is a convenient function to register the route with the
// context handler, which is the same as b.Method(http.MethodPost).Handler(handler).
func (b RouteBuilder) POSTContextWithError(handler reqresp.HandlerWithError) RouteBuilder {
	return b.Method(http.MethodPost).Handler(handler)
}

// DELETEContextWithError is a convenient function to register the route with the
// context handler, which is the same as b.Method(http.MethodDelete).Handler(handler).
func (b RouteBuilder) DELETEContextWithError(handler reqresp.HandlerWithError) RouteBuilder {
	return b.Method(http.MethodDelete).Handler(handler)
}

// PATCHContextWithError is a convenient function to register the route with the
// context handler, which is the same as b.Method(http.MethodPatch).Handler(handler).
func (b RouteBuilder) PATCHContextWithError(handler reqresp.HandlerWithError) RouteBuilder {
	return b.Method(http.MethodPatch).Handler(handler)
}

// HEADContextWithError is a convenient function to register the route with the
// context handler, which is the same as b.Method(http.MethodHead).Handler(handler).
func (b RouteBuilder) HEADContextWithError(handler reqresp.HandlerWithError) RouteBuilder {
	return b.Method(http.MethodHead).Handler(handler)
}

// OPTIONSContextWithError is a convenient function to register the route with the
// context handler, which is the same as b.Method(http.MethodOptions).Handler(handler).
func (b RouteBuilder) OPTIONSContextWithError(handler reqresp.HandlerWithError) RouteBuilder {
	return b.Method(http.MethodOptions).Handler(handler)
}
