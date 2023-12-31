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

package context

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/go-apiserver/http/reqresp"
)

func TestContext(t *testing.T) {
	c := reqresp.AcquireContext()
	r := new(http.Request)

	Context(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch reqresp.GetContext(r.Context()) {
		case nil:
			t.Errorf("expect to get a new context, but got nil")
		case c:
			t.Errorf("expect to get a new context, but got an old")
		}
	})).ServeHTTP(httptest.NewRecorder(), r)

	Context(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch reqresp.GetContext(r.Context()) {
		case c:
		case nil:
			t.Errorf("expect to get a context, but got nil")
		default:
			t.Errorf("expect to get an old context, but got a new")
		}
	})).ServeHTTP(httptest.NewRecorder(), r.WithContext(reqresp.SetContext(r.Context(), c)))
}
