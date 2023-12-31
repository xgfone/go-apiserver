// Copyright 2024 xgfone
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

package code

import (
	"net/http"

	"github.com/xgfone/go-apiserver/http/reqresp"
)

// StatusCode tries to inspect e.Ctx to the status code.
// If failed, return 500 instead.
func (e Error[T]) StatusCode() int {
	if code, ok := e.Ctx.(int); ok && code >= 100 && code < 600 {
		return code
	}
	return 500
}

// ServeHTTPWithCode stores the status code to e.Ctx and calls ServeHTTP.
func (e Error[T]) ServeHTTPWithCode(w http.ResponseWriter, r *http.Request, code int) {
	e.WithCtx(code).ServeHTTP(w, r)
}

// ServeHTTP implements the interface http.Handler.
func (e Error[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if c := reqresp.GetContext(r.Context()); c != nil {
		w = c
	}
	e.Respond(w)
}
