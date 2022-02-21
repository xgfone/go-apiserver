// Copyright 2022 xgfone
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
	"time"

	mw "github.com/xgfone/go-apiserver/http/middleware"
	"github.com/xgfone/go-apiserver/log"
)

// Logger returns a new http handler middleware to log the http request.
func Logger(priority int) mw.Middleware {
	return mw.NewMiddleware("logger", priority, func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			h.ServeHTTP(w, r)
			cost := time.Since(start)

			c := GetContext(r)

			level := log.LvlInfo
			if c.Err != nil {
				level = log.LvlError
			}

			if !log.Enabled(level) {
				return
			}

			kvs := make([]interface{}, 0, 12)
			kvs = append(kvs,
				"addr", r.RemoteAddr,
				"method", r.Method,
				"action", c.Action,
				"start", start.Unix(),
				"cost", cost)

			if c.Err != nil {
				kvs = append(kvs, "err", c.Err)
			}

			log.Log(level, 0, "log action request", kvs...)
		})
	})
}
