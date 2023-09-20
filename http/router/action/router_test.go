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

package action

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"

	"github.com/xgfone/go-apiserver/http/reqresp"
)

func TestActionRoute(t *testing.T) {
	router := NewRouter()
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(204)
	})

	router.RegisterContext("Test1", func(c *reqresp.Context) { c.WriteHeader(201) })
	router.RegisterContextWithError("Test2", func(c *reqresp.Context) error { c.WriteHeader(202); return nil })

	if actions := router.Actions(); len(actions) != 2 {
		t.Errorf("expect '%d' actions, but got '%d': %v", 2, len(actions), actions)
	} else {
		sort.Strings(actions)
		expects := []string{"Test1", "Test2"}
		if !reflect.DeepEqual(actions, expects) {
			t.Errorf("expect actions %v, but got %v", expects, actions)
		}
	}
	if n := len(router.Handlers()); n != 2 {
		t.Errorf("expect %d action handlers, but got %d", 2, n)
	}

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://127.0.0.1", nil)
	req.Header.Set(HeaderAction, "Test1")
	router.ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Errorf("expect the status code '%d', but got '%d'", rec.Code, 201)
	}

	rec = httptest.NewRecorder()
	req.Header.Set(HeaderAction, "Test2")
	router.ServeHTTP(rec, req)
	if rec.Code != 202 {
		t.Errorf("expect the status code '%d', but got '%d'", rec.Code, 202)
	}

	req.Header.Set(HeaderAction, "Test")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != 204 {
		t.Errorf("expect the status code '%d', but got '%d'", rec.Code, 204)
	}
}
