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
	"github.com/xgfone/go-defaults"
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

// Logger is used to wrap a http handler and return a new http handler,
// which logs the http request.
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
		} else if rw, ok := w.(reqresp.ResponseWriter); ok {
			code = rw.StatusCode()
		}

		attrs := getattrs()
		defer putattrs(attrs)

		attrs.Append(
			slog.String("raddr", r.RemoteAddr),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)

		if reqid := defaults.GetRequestID(ctx, r); reqid != "" {
			attrs.Append(slog.String("reqid", reqid))
		}
		if action != "" {
			attrs.Append(slog.String("action", action))
		}

		attrs.Append(
			slog.Int("code", code),
			slog.Int64("start", start.Unix()),
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

func getStacks(err error) []string {
	type stack interface {
		Stacks() []string
	}

	var s stack
	if errors.As(err, &s) {
		return s.Stacks()
	}
	return nil
}

type attrswrapper struct{ Attrs []slog.Attr }

func (w *attrswrapper) Reset()                { ; clear(w.Attrs); w.Attrs = w.Attrs[:0] }
func (w *attrswrapper) Append(a ...slog.Attr) { w.Attrs = append(w.Attrs, a...) }

var attrspool = &sync.Pool{New: func() interface{} {
	return &attrswrapper{Attrs: make([]slog.Attr, 0, 36)}
}}

func getattrs() *attrswrapper  { return attrspool.Get().(*attrswrapper) }
func putattrs(w *attrswrapper) { w.Reset(); attrspool.Put(w) }
