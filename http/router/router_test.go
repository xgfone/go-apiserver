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
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ghttp "github.com/xgfone/go-apiserver/http"
	"github.com/xgfone/go-apiserver/http/matcher"
)

func TestPriorityRoute(t *testing.T) {
	router := NewRouter()

	hostMatcher, _ := matcher.Host("127.0.0.1")
	router.Name("route1").Match(hostMatcher).
		HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(201)
			rw.Write([]byte(`route1`))
		})

	router.Rule("Host(`127.0.0.1`) && Method(`GET`)").Name("route2").
		HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(202)
			rw.Write([]byte(`route2`))
		})

	routes := router.GetRoutes()
	if _len := len(routes); _len != 2 {
		t.Errorf("expect %d routes, but got %d", 2, _len)
	} else {
		names := []string{"route2", "route1"}
		for i := 0; i < 2; i++ {
			if names[i] != routes[i].Name {
				t.Errorf("expect the route named '%s', but got '%s'",
					names[i], routes[i].Name)
			}
		}
	}

	req, _ := http.NewRequest("GET", "http://127.0.0.1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != 202 {
		t.Errorf("expect the status code '%d', but got '%d'", 202, rec.Code)
	} else if body := rec.Body.String(); body != "route2" {
		t.Errorf("expect the body '%s', but got '%s'", "route2", body)
	}

	router.DelRoute("route1")
	if route, ok := router.GetRoute("route1"); ok {
		t.Errorf("unexpected the route named '%s': '%s'", "route1", route.Name)
	}

	router.DelRoutes("route2")
	routes = router.GetRoutes()
	for _, route := range routes {
		t.Errorf("unexpected the route named '%s'", route.Name)
	}
}

func logMiddleware(buf *bytes.Buffer, name string) Middleware {
	return ghttp.NewMiddleware(name, func(h http.Handler) http.Handler {
		return ghttp.WrapHandler(h,
			func(h http.Handler, rw http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(buf, "middleware '%s' before\n", name)
				h.ServeHTTP(rw, r)
				fmt.Fprintf(buf, "middleware '%s' after\n", name)
			})
	})
}

func TestRouteMiddleware(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	handler := func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(200)
		buf.WriteString("handler\n")
	}

	router := NewRouter()
	router.Global(logMiddleware(buf, "log1"), logMiddleware(buf, "log2"))
	router.Use(logMiddleware(buf, "log3"), logMiddleware(buf, "log4"))
	router.Rule("Host(`127.0.0.1`) && Method(`GET`)").Name("route").HandlerFunc(handler)

	req, _ := http.NewRequest("GET", "http://127.0.0.1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expect the status code '%d', but got '%d'", 200, rec.Code)
	}

	expects := []string{
		"middleware 'log1' before",
		"middleware 'log2' before",
		"middleware 'log3' before",
		"middleware 'log4' before",
		"handler",
		"middleware 'log4' after",
		"middleware 'log3' after",
		"middleware 'log2' after",
		"middleware 'log1' after",
		"",
	}
	results := strings.Split(buf.String(), "\n")
	if len(results) != len(expects) {
		t.Errorf("expect %d lines, but got %d", len(expects), len(results))
	} else {
		for i := 0; i < len(results); i++ {
			if results[i] != expects[i] {
				t.Errorf("expect '%s', but got '%s'", expects[i], results[i])
			}
		}
	}

	buf.Reset()
	router.UseCancel("log3")
	router.GlobalCancel("log1")
	router.ServeHTTP(rec, req)

	expects = []string{
		"middleware 'log2' before",
		"middleware 'log3' before",
		"middleware 'log4' before",
		"handler",
		"middleware 'log4' after",
		"middleware 'log3' after",
		"middleware 'log2' after",
		"",
	}
	results = strings.Split(buf.String(), "\n")
	if len(results) != len(expects) {
		t.Errorf("expect %d lines, but got %d: %v", len(expects), len(results), results)
	} else {
		for i := 0; i < len(results); i++ {
			if results[i] != expects[i] {
				t.Errorf("expect '%s', but got '%s'", expects[i], results[i])
			}
		}
	}

	buf.Reset()
	router.UseCancel("log3")
	route, _ := router.GetRoute("route")
	route.Handler = ghttp.UnwrapHandler(route.Handler)
	router.UpdateRoutes(route)
	router.ServeHTTP(rec, req)

	expects = []string{
		"middleware 'log2' before",
		"middleware 'log4' before",
		"handler",
		"middleware 'log4' after",
		"middleware 'log2' after",
		"",
	}
	results = strings.Split(buf.String(), "\n")
	if len(results) != len(expects) {
		t.Errorf("expect %d lines, but got %d: %v", len(expects), len(results), results)
	} else {
		for i := 0; i < len(results); i++ {
			if results[i] != expects[i] {
				t.Errorf("expect '%s', but got '%s'", expects[i], results[i])
			}
		}
	}
}
