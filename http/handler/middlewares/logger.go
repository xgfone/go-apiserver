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
	"time"

	"github.com/xgfone/go-apiserver/http/handler"
	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/log"
)

// Logger returns a new http handler middleware to log the http request.
func Logger() handler.Middleware {
	return handler.NewMiddleware("logger", func(h http.Handler) http.Handler {
		return handler.WrapHandler(h, func(h http.Handler, w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			h.ServeHTTP(w, r)
			cost := time.Since(start)

			var err error
			if c := reqresp.GetContext(r); c != nil {
				err = c.Err
			}

			level := log.LvlInfo
			if err != nil {
				level = log.LvlError
			}

			if !log.Enabled(level) {
				return
			}

			kvs := make([]interface{}, 0, 14)
			kvs = append(kvs,
				"addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.Path,
				"start", start.Unix(),
				"cost", cost)

			if rw, ok := w.(reqresp.ResponseWriter); ok {
				kvs = append(kvs, "code", rw.StatusCode())
			}

			if err != nil {
				kvs = append(kvs, "err", err)
			}

			log.Log(level, 0, "log http request", kvs...)
		})
	})
}
