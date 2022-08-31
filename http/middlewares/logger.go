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

	"github.com/xgfone/go-apiserver/http/header"
	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/middleware"
	"github.com/xgfone/go-apiserver/middleware/logger"
	"github.com/xgfone/go-apiserver/tools/pool"
	"github.com/xgfone/go-apiserver/tools/rawjson"
)

// LoggerHandler is used to handle the extra logs.
type LoggerHandler func(http.ResponseWriter, *http.Request, []interface{}) []interface{}

// Logger is equal to LoggerWithOptions(priority, nil).
func Logger(priority int) middleware.Middleware { return LoggerWithOptions(priority, nil) }

// LoggerWithOptions returns a new common http handler middleware to log the http request.
func LoggerWithOptions(priority int, handler LoggerHandler, options ...logger.Option) middleware.Middleware {
	var config logger.Config
	for _, opt := range options {
		opt(&config)
	}

	return middleware.NewMiddleware("logger", priority, func(h interface{}) interface{} {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logLevel := config.GetLogLevel()
			if !log.Enabled(logLevel) {
				h.(http.Handler).ServeHTTP(w, r)
				return
			}

			ctx := reqresp.GetContext(w, r)

			var reqBodyLen int
			var reqBodyData string
			logReqBodyLen := config.GetLogReqBodyLen()
			if logReqBodyLen > 0 && (r.ContentLength <= 0 || r.ContentLength <= int64(logReqBodyLen)) {
				reqBuf := pool.GetBuffer(logReqBodyLen)
				defer reqBuf.Release()

				_, err := io.CopyBuffer(reqBuf, r.Body, make([]byte, 1024))
				if err != nil {
					log.Error("fail to read the request body", "raddr", r.RemoteAddr,
						"method", r.Method, "path", r.RequestURI, "err", err)
				}

				reqBodyLen = reqBuf.Len()
				if reqBodyLen <= logReqBodyLen {
					reqBodyData = reqBuf.String()
				} else {
					logReqBodyLen = -1
				}

				defer resetReqBody(r, r.Body)
				r.Body = bufferCloser{Buffer: reqBuf.Buffer, Closer: r.Body}
			}

			var respBuf *pool.Buffer
			logRespBodyLen := config.GetLogRespBodyLen()
			if logRespBodyLen > 0 {
				respBuf = pool.GetBuffer(logRespBodyLen)
				defer respBuf.Release()

				rw := reqresp.NewResponseWriterWithWriteResponse(w,
					func(w http.ResponseWriter, b []byte) (int, error) {
						n, err := w.Write(b)
						if n > 0 {
							respBuf.Write(b[:n])
						}
						return n, err
					})

				if ctx != nil {
					ctx.ResponseWriter = rw
				}
				w = rw
			}

			start := time.Now()
			h.(http.Handler).ServeHTTP(w, r)
			cost := time.Since(start)

			var code int
			var err error
			if ctx != nil {
				code = ctx.StatusCode()
				err = ctx.Err
			} else if rw, ok := w.(reqresp.ResponseWriter); ok {
				code = rw.StatusCode()
			}

			ikvs := pool.GetInterfaces(32)
			kvs := ikvs.Interfaces
			kvs = append(kvs,
				"raddr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.Path,
				"code", code,
				"start", start.Unix(),
				"cost", cost,
			)

			if handler != nil {
				kvs = handler(w, r, kvs)
			}

			if config.GetLogReqHeaders() {
				kvs = append(kvs, "reqheaders", r.Header)
			}

			if reqBodyLen <= logReqBodyLen {
				kvs = append(kvs, "reqbodylen", reqBodyLen)

				if header.ContentType(r.Header) == header.MIMEApplicationJSON {
					// (xgfone): We needs to check whether reqbody is a valid raw json string??
					kvs = append(kvs, "reqbodydata", rawjson.RawString(reqBodyData))
				} else {
					kvs = append(kvs, "reqbodydata", reqBodyData)
				}
			}

			if config.GetLogRespHeaders() {
				kvs = append(kvs, "respheaders", w.Header())
			}

			if respBuf != nil && respBuf.Len() <= logRespBodyLen {
				kvs = append(kvs, "respbodylen", respBuf.Len())

				if header.ContentType(w.Header()) == header.MIMEApplicationJSON {
					// (xgfone): We needs to check whether respbody is a valid raw json string??
					kvs = append(kvs, "respbodydata", rawjson.RawString(respBuf.String()))
				} else {
					kvs = append(kvs, "respbodydata", respBuf.String())
				}
			}

			if err != nil {
				kvs = append(kvs, "err", err)
			}

			log.Log(logLevel, 0, "log http request", kvs...)
			ikvs.Interfaces = kvs
			ikvs.Release()
		})
	})
}

type bufferCloser struct {
	*bytes.Buffer
	io.Closer
}

func resetReqBody(r *http.Request, body io.ReadCloser) { r.Body = body }
