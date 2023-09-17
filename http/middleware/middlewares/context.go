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

package middlewares

import (
	"net/http"

	"github.com/xgfone/go-apiserver/http/reqresp"
)

// Context is used to wrap a http handler and return a new http handler,
// which will allocate a context and put it into the http request,
// then release it after handling the http request.
func Context(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c := reqresp.GetContext(w, r); c == nil {
			c = reqresp.DefaultContextAllocator.Acquire()
			c.ResponseWriter = reqresp.NewResponseWriter(w)
			c.Request = reqresp.SetContext(r, c)
			defer reqresp.DefaultContextAllocator.Release(c)
			next.ServeHTTP(c.ResponseWriter, c.Request)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
