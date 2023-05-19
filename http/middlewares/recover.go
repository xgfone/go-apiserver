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

	"github.com/xgfone/go-apiserver/helper"
	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/middleware"
	"github.com/xgfone/go-apiserver/result"
)

// DefaultPanicHandler is the global default panic handler.
var DefaultPanicHandler PanicHandler

// PanicHandler is used to handle the panic.
//
// If returning true, no longer continue to do something.
// Or, do extra something, for example, log the panic, etc.
type PanicHandler func(w http.ResponseWriter, r *http.Request, recover interface{})

// Recover is equal to RecoverWithHandler(priority, nil).
func Recover(priority int) middleware.Middleware {
	return RecoverWithHandler(priority, nil)
}

// RecoverWithHandler returns a new http handler middleware with the panic
// handler, which is used to wrap and recover the panic.
//
// If handler is nil, use the global DefaultPanicHandler instead.
// If DefaultPanicHandler is also nil, use the inner default handler,
// which will respond with the status code 500.
func RecoverWithHandler(priority int, handler PanicHandler) middleware.Middleware {
	return middleware.NewMiddleware("recover", priority, func(h interface{}) interface{} {
		next := h.(http.Handler)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer wrapPanic(w, r, handler)
			next.ServeHTTP(w, r)
		})
	})
}

func wrapPanic(w http.ResponseWriter, r *http.Request, handler PanicHandler) {
	if e := recover(); e != nil {
		if handler != nil {
			handler(w, r, e)
		} else if DefaultPanicHandler != nil {
			DefaultPanicHandler(w, r, e)
		} else {
			defaultHandler(w, r, e)
		}
	}
}

func defaultHandler(w http.ResponseWriter, r *http.Request, recover interface{}) {
	var rw reqresp.ResponseWriter
	c := reqresp.GetContext(w, r)
	if c != nil {
		rw = c.ResponseWriter
		c.Err = panicError{panics: recover, stacks: helper.GetCallStack(5)}
	} else if _rw, ok := w.(reqresp.ResponseWriter); ok {
		rw = _rw
	} else {
		return
	}

	if !rw.WroteHeader() {
		if c == nil || c.Action == "" {
			rw.WriteHeader(500)
			fmt.Fprint(rw, recover)
		} else {
			var rerr result.Error
			switch err := recover.(type) {
			case result.Error:
				rerr = err

			case error:
				rerr = result.ErrInternalServerError.WithError(err)

			default:
				rerr = result.ErrInternalServerError.WithMessage("%v", err)
			}

			c.Respond(result.Response{Error: rerr})
		}
	}
}

type panicError struct {
	panics interface{}
	stacks []string
}

func (e panicError) Error() string    { return fmt.Sprintf("wrap a panic: %v", e.panics) }
func (e panicError) Stacks() []string { return e.stacks }
