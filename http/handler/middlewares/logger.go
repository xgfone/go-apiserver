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
	glog "github.com/xgfone/go-log"
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

			var logger *glog.Emitter
			if err == nil {
				logger = log.Info()
			} else {
				logger = log.Error().Err(err)
			}

			if rw, ok := w.(reqresp.ResponseWriter); ok {
				logger = logger.Int("code", rw.StatusCode())
			}

			logger.Str("addr", r.RemoteAddr).Str("method", r.Method).
				Str("uri", r.RequestURI).Int64("start", start.Unix()).
				Duration("cost", cost).Printf("log http request")
		})
	})
}
