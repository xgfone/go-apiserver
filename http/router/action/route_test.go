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

package action

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/internal/test"
)

func TestActionRoute(t *testing.T) {
	router := NewRouter()
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqresp.GetContext(w, r).WriteHeader(204)
	})

	router.RegisterContextFunc("Test1", func(c *reqresp.Context) { c.WriteHeader(201) })
	router.RegisterContextFunc("Test2", func(c *reqresp.Context) { c.WriteHeader(202) })

	if actions := router.GetActions(); len(actions) != 2 {
		t.Errorf("expect '%d' actions, but got '%d': %v", 2, len(actions), actions)
	} else {
		test.InStrings(t, "TestActionRoute", actions, []string{"Test1", "Test2"})
	}

	req, _ := http.NewRequest("GET", "http://127.0.0.1", nil)
	req.Header.Set(HeaderAction, "Test1")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	test.CheckStatusCode(t, "TestActionRoute", rec.Code, 201)

	req.Header.Set(HeaderAction, "Test")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	test.CheckStatusCode(t, "TestActionRoute", rec.Code, 204)

	router.Unregister("Test1")
	if actions := router.GetActions(); len(actions) != 1 {
		t.Errorf("expect '%d' actions, but got '%d': %v", 1, len(actions), actions)
	} else if actions[0] != "Test2" {
		t.Errorf("expect the action '%s', but got '%s'", "Test2", actions[0])
	}
}
