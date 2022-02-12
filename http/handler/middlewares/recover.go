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
	"fmt"
	"net/http"

	"github.com/xgfone/go-apiserver/http/handler"
	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/log"
)

// Recover returns a new http handler middleware, which is used to wrap
// and recover the panic.
func Recover() handler.Middleware {
	return handler.NewMiddleware("recover", func(h http.Handler) http.Handler {
		return handler.WrapHandler(h, func(h http.Handler, w http.ResponseWriter, r *http.Request) {
			defer wrapPanic(w, r)
			h.ServeHTTP(w, r)
		})
	})
}

func wrapPanic(w http.ResponseWriter, r *http.Request) {
	if e := recover(); e != nil {
		if c := reqresp.GetContext(r); c == nil || c.Err != nil {
			log.Error().Str("addr", r.RemoteAddr).Str("method", r.Method).
				Str("uri", r.RequestURI).Kv("panic", e).Printf("wrap a panic")
		} else if err, ok := c.Err.(error); ok {
			c.Err = fmt.Errorf("panic: %w", err)
		} else {
			c.Err = fmt.Errorf("panic: %v", e)
		}

		if rw, ok := w.(reqresp.ResponseWriter); ok && !rw.WroteHeader() {
			rw.WriteHeader(500)
		}
	}
}