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
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/xgfone/go-apiserver/helper"
	"github.com/xgfone/go-apiserver/log"
	mw "github.com/xgfone/go-apiserver/middleware"
)

func wrapPanic(w http.ResponseWriter, r *http.Request) {
	if err := recover(); err != nil {
		stacks := helper.GetCallStack(helper.RecoverStackSkip)
		log.Error("wrap a panic", "addr", r.RemoteAddr, "method", r.Method,
			"uri", r.RequestURI, "panic", err, "stacks", stacks)

		switch e := err.(type) {
		case Error:
			GetContext(w, r).Err = e

		case string:
			GetContext(w, r).Err = ErrInternalServerError.WithMessage(e)

		case error:
			GetContext(w, r).Err = ErrInternalServerError.WithMessage(e.Error())

		default:
			GetContext(w, r).Err = ErrInternalServerError.WithMessage(fmt.Sprint(e))
		}
	}
}

// Recover returns a new http handler middleware to wrap the panic as an error
// and recover the handling process.
func Recover(priority int) mw.Middleware {
	return mw.NewMiddleware("recover", priority, func(h interface{}) interface{} {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			defer wrapPanic(rw, r)
			h.(http.Handler).ServeHTTP(rw, r)
		})
	})
}

// LoggerConfig is used to configure the logger middleware.
type LoggerConfig struct {
	Priority   int
	LogLevel   int
	LogReqBody bool
}

// Logger is a convenient logger middleware, which is equal to
//   LoggerWithConfig(LoggerConfig{Priority: priority, LogLevel: log.LvlInfo})
func Logger(priority int) mw.Middleware {
	return LoggerWithConfig(LoggerConfig{Priority: priority, LogLevel: log.LvlInfo})
}

// LoggerWithConfig returns a new http handler middleware to log the http request.
func LoggerWithConfig(c LoggerConfig) mw.Middleware {
	return mw.NewMiddleware("logger", c.Priority, func(h interface{}) interface{} {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !log.Enabled(c.LogLevel) {
				h.(http.Handler).ServeHTTP(w, r)
				return
			}

			var reqbody string
			if c.LogReqBody {
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

			ctx := GetContext(w, r)
			kvs := make([]interface{}, 0, 16)
			kvs = append(kvs,
				"addr", r.RemoteAddr,
				"code", ctx.StatusCode(),
				"method", r.Method,
				"action", ctx.Action,
				"start", start.Unix(),
				"cost", cost,
			)

			if c.LogReqBody {
				kvs = append(kvs, "reqbody", reqbody)
			}

			if ctx.Err != nil {
				kvs = append(kvs, "err", ctx.Err)
			}

			log.Log(c.LogLevel, 0, "log http request", kvs...)
		})
	})
}

type bufferCloser struct {
	*bytes.Buffer
	io.Closer
}
