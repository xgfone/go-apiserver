// Copyright 2021~2023 xgfone
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
	"sync"
	"time"

	"github.com/xgfone/go-apiserver/http/header"
	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/internal/pools"
	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/middleware"
	"github.com/xgfone/go-apiserver/middleware/logger"
	"github.com/xgfone/go-defaults"
)

// LogKvsAppender is used to append the extra log key-value contexts.
type LogKvsAppender func(http.ResponseWriter, *http.Request, []interface{}) []interface{}

// Logger is equal to LoggerWithOptions(priority, nil).
func Logger(priority int) middleware.Middleware { return LoggerWithOptions(priority, nil) }

// LoggerWithOptions returns a new common http handler middleware to log the http request.
func LoggerWithOptions(priority int, appender LogKvsAppender, options ...logger.Option) middleware.Middleware {
	var config logger.Config
	for _, opt := range options {
		opt(&config)
	}

	return middleware.NewMiddleware("logger", priority, func(h interface{}) interface{} {
		next := h.(http.Handler)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if !log.Enabled(ctx, log.LevelInfo) || !config.GetLogReq(ctx) {
				next.ServeHTTP(w, r)
				return
			}

			c := reqresp.GetContext(w, r)

			var reqBodyLen int
			var reqBodyData string
			logReqBodyLen := config.GetLogReqBodyLen(ctx)
			if logReqBodyLen > 0 {
				if r.ContentLength < 0 || r.ContentLength > int64(logReqBodyLen) {
					logReqBodyLen = -1
				} else {
					pool, reqBuf := pools.GetBuffer(logReqBodyLen)
					defer pools.PutBuffer(pool, reqBuf)

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
					r.Body = bufferCloser{Buffer: reqBuf, Closer: r.Body}
				}
			}

			var respBuf *bytes.Buffer
			logRespBodyLen := config.GetLogRespBodyLen(ctx)
			if logRespBodyLen > 0 {
				var pool *sync.Pool
				pool, respBuf = pools.GetBuffer(logRespBodyLen)
				defer pools.PutBuffer(pool, respBuf)

				rw := reqresp.NewResponseWriter(w, reqresp.WriteWithResponse(
					func(w http.ResponseWriter, b []byte) (int, error) {
						n, err := w.Write(b)
						if n > 0 {
							respBuf.Write(b[:n])
						}
						return n, err
					}), reqresp.DisableReaderFrom())

				if c != nil {
					c.ResponseWriter = rw
				}
				w = rw
			}

			start := time.Now()
			next.ServeHTTP(w, r)
			cost := time.Since(start)

			var code int
			var err error
			if c != nil {
				code = c.StatusCode()
				err = c.Err
			} else if rw, ok := w.(reqresp.ResponseWriter); ok {
				code = rw.StatusCode()
			}

			ipool, ikvs := pools.GetInterfaces(32)
			kvs := ikvs.Interfaces
			kvs = append(kvs,
				"raddr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.Path,
				"code", code,
				"start", start.Unix(),
				"cost", cost.String(),
			)

			if c != nil && c.Action != "" {
				kvs = append(kvs, "action", c.Action)
			}

			if reqid := defaults.GetRequestID(ctx, r); reqid != "" {
				kvs = append(kvs, "reqid", reqid)
			}

			if appender != nil {
				kvs = appender(w, r, kvs)
			}

			if config.GetLogReqQuery(ctx) {
				kvs = append(kvs, "query", r.URL.RawQuery)
			}

			if config.GetLogReqHeaders(ctx) {
				kvs = append(kvs, "reqheaders", r.Header)
			}

			if logReqBodyLen > 0 {
				kvs = append(kvs, "reqbodylen", reqBodyLen)
				if reqBodyLen <= logReqBodyLen {
					if header.ContentType(r.Header) == header.MIMEApplicationJSON {
						// (xgfone): We needs to check whether reqbody is a valid raw json string??
						kvs = append(kvs, "reqbodydata", rawString(reqBodyData))
					} else {
						kvs = append(kvs, "reqbodydata", reqBodyData)
					}
				}
			}

			if config.GetLogRespHeaders(ctx) {
				kvs = append(kvs, "respheaders", w.Header())
			}

			if respBuf != nil {
				kvs = append(kvs, "respbodylen", respBuf.Len())
				if respBuf.Len() <= logRespBodyLen {
					if header.ContentType(w.Header()) == header.MIMEApplicationJSON {
						// (xgfone): We needs to check whether respbody is a valid raw json string??
						kvs = append(kvs, "respbodydata", rawString(respBuf.String()))
					} else {
						kvs = append(kvs, "respbodydata", respBuf.String())
					}
				}
			}

			if err != nil {
				kvs = append(kvs, "err", err)
				if se, ok := err.(interface{ Stacks() []string }); ok {
					kvs = append(kvs, "stacks", se.Stacks())
				}
			}

			log.Info("log http request", kvs...)
			ikvs.Interfaces = kvs
			pools.PutInterfaces(ipool, ikvs)
		})
	})
}

type bufferCloser struct {
	*bytes.Buffer
	io.Closer
}

func resetReqBody(r *http.Request, body io.ReadCloser) { r.Body = body }
