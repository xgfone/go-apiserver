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
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/log"
	mw "github.com/xgfone/go-apiserver/middleware"
)

// Logger is a convenient logger middleware, which is equal to
//   LoggerWithConfig(middleware.NewLoggerConfig(priority, log.LvlInfo, false))
func Logger(priority int) mw.Middleware {
	return LoggerWithConfig(mw.NewLoggerConfig(priority, log.LvlInfo, false))
}

// LoggerWithConfig returns a new http handler middleware to log the http request.
func LoggerWithConfig(c mw.LoggerConfig) mw.Middleware {
	return mw.NewMiddleware("logger", c.Priority, func(h interface{}) interface{} {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logLevel := c.GetLogLevel()
			if !log.Enabled(logLevel) {
				h.(http.Handler).ServeHTTP(w, r)
				return
			}

			var reqbody string
			logReqBody := c.GetLogReqBody()
			if logReqBody {
				reqbuf := bytes.NewBuffer(nil)
				if r.ContentLength > 0 {
					reqbuf.Grow(int(r.ContentLength))
					io.CopyN(reqbuf, r.Body, r.ContentLength)
				} else {
					io.CopyBuffer(reqbuf, r.Body, make([]byte, 2048))
				}
				reqbody = reqbuf.String()
				r.Body = bufferCloser{Buffer: reqbuf, Closer: r.Body}
			}

			start := time.Now()
			h.(http.Handler).ServeHTTP(w, r)
			cost := time.Since(start)

			var code int
			var err error
			if ctx := reqresp.GetContext(w, r); ctx != nil {
				code = ctx.StatusCode()
				err = ctx.Err
			} else if rw, ok := w.(reqresp.ResponseWriter); ok {
				code = rw.StatusCode()
			}

			kvs := make([]interface{}, 0, 16)
			kvs = append(kvs,
				"addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.Path,
				"code", code,
				"start", start.Unix(),
				"cost", cost,
			)

			if logReqBody {
				kvs = append(kvs, "reqbody", reqbody)
			}

			if err != nil {
				kvs = append(kvs, "err", err)
			}

			log.Log(logLevel, 0, "log http request", kvs...)
		})
	})
}

type bufferCloser struct {
	*bytes.Buffer
	io.Closer
}
