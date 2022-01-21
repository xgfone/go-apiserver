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

package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/go-apiserver/http/reqresp"
)

func TestDefaultMiddlewares(t *testing.T) {
	handler := func(rw http.ResponseWriter, r *http.Request) {
		if reqresp.GetContext(r) == nil {
			t.Errorf("no the request context")
		}

		if _, ok := rw.(reqresp.ResponseWriter); !ok {
			t.Errorf("http.ResponseWriter is not reqresp.ResponseWriter")
		}
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	DefaultMiddlewares.Handler(http.HandlerFunc(handler)).ServeHTTP(rec, req)
}
