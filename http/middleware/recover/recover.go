// Copyright 2023~2024 xgfone
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

// Package recover provides a recover middleware to wrap and log the panic.
package recover

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/result"
	"github.com/xgfone/go-apiserver/result/codeint"
	"github.com/xgfone/go-toolkit/runtimex"
)

// Recover is a http handler middleware to recover the panic if occurring.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer wrappanic(w, r)
		next.ServeHTTP(w, r)
	})
}

func wrappanic(w http.ResponseWriter, r *http.Request) {
	v := recover()
	if v == nil {
		return
	}

	stacks := runtimex.Stacks(2)
	if c := reqresp.GetContext(r.Context()); c != nil {
		c.AppendError(panicerror{stacks: stacks, panicv: v})
		if !c.ResponseWriter.WroteHeader() {
			c.ResponseWriter.Header().Set("X-Panic", "1")
			c.ResponseWriter.WriteHeader(500)
		}
		return
	}

	slog.Error("wrap a panic", slog.Any("panic", v), slog.Any("stacks", stacks))
	if !reqresp.WroteHeader(w) {
		w.Header().Set("X-Panic", "1")
		err := codeint.ErrInternalServerError.WithMessage("panic")
		reqresp.DefaultRespond(w, r, result.Err(err))
	}
}

type panicerror struct {
	stacks []runtimex.Frame
	panicv any
}

func (e panicerror) Error() string            { return fmt.Sprintf("panic: %v", e.panicv) }
func (e panicerror) Stacks() []runtimex.Frame { return e.stacks }
