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

	mw "github.com/xgfone/go-apiserver/http/middleware"
	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/log"
)

// Recover returns a new http handler middleware, which is used to wrap
// and recover the panic.
func Recover(priority int) mw.Middleware {
	return mw.NewMiddleware("recover", priority, func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer wrapPanic(w, r)
			h.ServeHTTP(w, r)
		})
	})
}

func wrapPanic(w http.ResponseWriter, r *http.Request) {
	if e := recover(); e != nil {
		if c := reqresp.GetContext(r); c == nil {
			log.Error("wrap a panic", "addr", r.RemoteAddr, "method", r.Method,
				"uri", r.RequestURI, "panic", e)

			if rw, ok := w.(reqresp.ResponseWriter); ok && !rw.WroteHeader() {
				rw.WriteHeader(500)
			}
		} else if err, ok := e.(error); ok {
			c.Err = fmt.Errorf("panic: %w", err)
		} else {
			c.Err = fmt.Errorf("panic: %v", e)
		}
	}
}
