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
	"fmt"
	"net/http"
	"time"

	"github.com/xgfone/go-apiserver/log"
	mw "github.com/xgfone/go-apiserver/middleware"
)

func wrapPanic(w http.ResponseWriter, r *http.Request) {
	if err := recover(); err != nil {
		stacks := log.GetCallStack(log.RecoverStackSkip)
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

// Logger returns a new http handler middleware to log the http request.
func Logger(priority int) mw.Middleware {
	return mw.NewMiddleware("logger", priority, func(h interface{}) interface{} {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			h.(http.Handler).ServeHTTP(w, r)
			cost := time.Since(start)

			c := GetContext(w, r)
			if !c.WroteHeader() {
				if c.Err == nil {
					c.Success(nil)
				} else {
					c.Failure(c.Err)
				}
			}

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
				"code", c.StatusCode(),
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
