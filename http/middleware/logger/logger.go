// Copyright 2023 xgfone
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

// Package logger provides a logger middleware to log the http request.
package logger

import (
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-toolkit/runtimex"
)

var (
	// Collect is used to collect the extra key-value attributes if set.
	//
	// Default: nil
	Collect func(w http.ResponseWriter, r *http.Request, append func(...slog.Attr))

	// Enabled is used to decide whether to log the request if set.
	//
	// Default: nil
	Enabled func(*http.Request) bool
)

// Logger is a http middleware to log the http request.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if (Enabled != nil && !Enabled(r)) || !slog.Default().Enabled(ctx, slog.LevelInfo) {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		next.ServeHTTP(w, r)
		cost := time.Since(start)

		var err error
		var code = 200
		var action string
		if c := reqresp.GetContext(r.Context()); c != nil {
			action = c.Action
			code = c.StatusCode()
			if c.Err != nil {
				err = c.Err
			}
		} else if rw, ok := w.(interface{ StatusCode() int }); ok {
			code = rw.StatusCode()
		}

		attrs := getattrs()
		defer putattrs(attrs)

		attrs.Append(
			slog.String("raddr", r.RemoteAddr),
			slog.String("method", r.Method),
			slog.String("host", r.Host),
			slog.String("path", r.URL.Path),
		)

		if reqid := r.Header.Get("X-Request-Id"); reqid != "" {
			attrs.Append(slog.String("reqid", reqid))
		}

		if action != "" {
			attrs.Append(slog.String("action", action))
		}

		attrs.Append(
			slog.Int("code", code),
			slog.String("cost", cost.String()),
		)

		if Collect != nil {
			Collect(w, r, attrs.Append)
		}

		if err != nil {
			attrs.Append(slog.String("err", err.Error()))
			if stacks := getStacks(err); stacks != nil {
				attrs.Append(slog.Any("stacks", stacks))
			}
		}

		slog.LogAttrs(r.Context(), slog.LevelInfo, "log http request", attrs.Attrs...)
	})
}

func getStacks(err error) any {
	type (
		stringstack interface {
			Stacks() []string
		}

		framestack interface {
			Stacks() []runtimex.Frame
		}
	)

	switch e := err.(type) {
	case framestack:
		return e.Stacks()

	case stringstack:
		return e.Stacks()

	default:
		var fstack framestack
		if errors.As(err, &fstack) {
			return fstack.Stacks()
		}

		var sstack stringstack
		if errors.As(err, &sstack) {
			return sstack.Stacks()
		}

		return nil
	}
}

type attrswrapper struct{ Attrs []slog.Attr }

func (w *attrswrapper) Reset()                { ; clear(w.Attrs); w.Attrs = w.Attrs[:0] }
func (w *attrswrapper) Append(a ...slog.Attr) { w.Attrs = append(w.Attrs, a...) }

var attrspool = &sync.Pool{New: func() any {
	return &attrswrapper{Attrs: make([]slog.Attr, 0, 36)}
}}

func getattrs() *attrswrapper  { return attrspool.Get().(*attrswrapper) }
func putattrs(w *attrswrapper) { w.Reset(); attrspool.Put(w) }
