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
	"net/http"
	"net/http/httptest"
	"testing"

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
