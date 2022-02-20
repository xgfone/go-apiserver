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

	mw "github.com/xgfone/go-apiserver/http/middleware"
	"github.com/xgfone/go-apiserver/http/reqresp"
)

// ResponseWriter returns a new http handler middleware, which converts
// http.ResponseWriter to the extended http ResponseWriter that supports
// to get the status code of the response.
func ResponseWriter(priority int) mw.Middleware {
	return mw.NewMiddleware("responsewriter", priority, func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(reqresp.NewResponseWriter(w), r)
		})
	})
}
