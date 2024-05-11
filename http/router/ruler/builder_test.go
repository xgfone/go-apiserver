// Copyright 2023~2024 xgfone
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
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/xgfone/go-apiserver/http/handler"
	"github.com/xgfone/go-apiserver/http/middleware"
	"github.com/xgfone/go-apiserver/http/reqresp"
)

func TestRouteBuilder(t *testing.T) {
	b := NewRouteBuilder(func(r Route) {
		desc := r.Matcher.(fmt.Stringer).String()
		if expect := r.Extra.(string); !strings.Contains(desc, expect) {
			t.Errorf("expect containing '%s', but got '%s'", expect, desc)
		}
		if expect := "path"; !strings.Contains(desc, expect) {
			t.Errorf("expect containing '%s', but got '%s'", expect, desc)
		}
		if expect := "localhost"; !strings.Contains(desc, expect) {
			t.Errorf("expect containing '%s', but got '%s'", expect, desc)
		}
	})

	b = b.Group("/prefix").Clone().Path("/path").Host("localhost").
		Use(middleware.MiddlewareFunc(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})
		}))

	if b.Prefix() != "/prefix" {
		t.Errorf("expec group prefix '%s', but got '%s'", "/prefix", b.Prefix())
	}

	b.Extra(http.MethodGet).GET(handler.Handler204)
	b.Extra(http.MethodPut).PUT(handler.Handler204)
	b.Extra(http.MethodPost).POST(handler.Handler204)
	b.Extra(http.MethodDelete).DELETE(handler.Handler204)
	b.Extra(http.MethodPatch).PATCH(handler.Handler204)
	b.Extra(http.MethodHead).HEAD(handler.Handler204)
	b.Extra(http.MethodOptions).OPTIONS(handler.Handler204)

	b.Extra(http.MethodGet).GETFunc(handler.Handler204.(http.HandlerFunc))
	b.Extra(http.MethodPut).PUTFunc(handler.Handler204.(http.HandlerFunc))
	b.Extra(http.MethodPost).POSTFunc(handler.Handler204.(http.HandlerFunc))
	b.Extra(http.MethodHead).HEADFunc(handler.Handler204.(http.HandlerFunc))
	b.Extra(http.MethodPatch).PATCHFunc(handler.Handler204.(http.HandlerFunc))
	b.Extra(http.MethodDelete).DELETEFunc(handler.Handler204.(http.HandlerFunc))
	b.Extra(http.MethodOptions).OPTIONSFunc(handler.Handler204.(http.HandlerFunc))

	chandler := func(c *reqresp.Context) {}
	b.Extra(http.MethodGet).GETContext(reqresp.Handler(chandler))
	b.Extra(http.MethodPut).PUTContext(reqresp.Handler(chandler))
	b.Extra(http.MethodPost).POSTContext(reqresp.Handler(chandler))
	b.Extra(http.MethodHead).HEADContext(reqresp.Handler(chandler))
	b.Extra(http.MethodPatch).PATCHContext(reqresp.Handler(chandler))
	b.Extra(http.MethodDelete).DELETEContext(reqresp.Handler(chandler))
	b.Extra(http.MethodOptions).OPTIONSContext(reqresp.Handler(chandler))

	cehandler := func(c *reqresp.Context) error { return nil }
	b.Extra(http.MethodGet).GETContextWithError(reqresp.HandlerWithError(cehandler))
	b.Extra(http.MethodPut).PUTContextWithError(reqresp.HandlerWithError(cehandler))
	b.Extra(http.MethodPost).POSTContextWithError(reqresp.HandlerWithError(cehandler))
	b.Extra(http.MethodHead).HEADContextWithError(reqresp.HandlerWithError(cehandler))
	b.Extra(http.MethodPatch).PATCHContextWithError(reqresp.HandlerWithError(cehandler))
	b.Extra(http.MethodDelete).DELETEContextWithError(reqresp.HandlerWithError(cehandler))
	b.Extra(http.MethodOptions).OPTIONSContextWithError(reqresp.HandlerWithError(cehandler))
}

func TestRouteBuilder_WrapRegister(t *testing.T) {
	var routes []Route
	b := NewRouteBuilder(func(route Route) { routes = append(routes, route) })
	b = b.WrapRegister(func(register func(Route), route Route) {
		route.Desc = "wrap"
		register(route)
	})

	b.Path("/path1").GETFunc(func(w http.ResponseWriter, r *http.Request) {})
	b.Path("/path2").GETFunc(func(w http.ResponseWriter, r *http.Request) {})

	if len(routes) != 2 {
		t.Fatalf("expect %d routes, but got %d", 2, len(routes))
	}

	for _, r := range routes {
		if r.Desc != "wrap" {
			t.Errorf("missing the flag desc 'wrap'")
		}
	}
}
