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

	"github.com/xgfone/go-apiserver/http/handler"
	"github.com/xgfone/go-apiserver/http/reqresp"
)

// Context returns a new http handler middleware, which will allocate
// a request context and put it into the http request, then release it after
// handling the http request.
func Context() handler.Middleware {
	return handler.NewMiddleware("context", func(h http.Handler) http.Handler {
		return handler.WrapHandler(h, func(h http.Handler, w http.ResponseWriter, r *http.Request) {
			if c, new := reqresp.GetOrNewContext(r); new {
				if rw, ok := w.(reqresp.ResponseWriter); ok {
					c.ResponseWriter = rw
				} else {
					c.ResponseWriter = reqresp.NewResponseWriter(w)
				}
				w = c.ResponseWriter
				r = reqresp.SetContext(r, c)
				defer reqresp.DefaultContextAllocator.Release(c)
			}
			h.ServeHTTP(w, r)
		})
	})
}
