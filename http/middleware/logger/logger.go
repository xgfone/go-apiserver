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
	"log/slog"
	"net/http"
	"time"

	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/internal/pools"
	"github.com/xgfone/go-defaults"
)

var (
	// Start is used to log the extra information.
	//
	// If returning nil, do not log any extra information.
	//
	// For the default implementation, it returns nil.
	Start func(http.ResponseWriter, *http.Request) Collector = defaultStart

	// Enabled is used to decide whether to log the request,
	//
	// For the default implementation, it returns true.
	Enabled func(*http.Request) bool = defaultEnabled
)

// Collector is used to collect the extra log key-value information.
//
// If the returned clean function is nil, it indicates not to need to clean any.
type Collector func(kvs []interface{}) (newkvs []interface{}, clean func())

func defaultStart(http.ResponseWriter, *http.Request) Collector { return nil }
func defaultEnabled(*http.Request) bool                         { return true }

// Logger is used to wrap a http handler and return a new http handler,
// which logs the http request.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if !Enabled(r) || !slog.Default().Enabled(ctx, slog.LevelInfo) {
			next.ServeHTTP(w, r)
			return
		}

		collect := Start(w, r)
		start := time.Now()
		next.ServeHTTP(w, r)
		cost := time.Since(start)

		var err error
		var code = 200
		var action string
		if c := reqresp.GetContext(r.Context()); c != nil {
			action = c.Action
			code = c.StatusCode()
			if c.LogErr != nil {
				err = c.LogErr
			} else {
				err = c.RespErr
			}
		} else if rw, ok := w.(reqresp.ResponseWriter); ok {
			code = rw.StatusCode()
		}

		ipool, ikvs := pools.GetInterfaces(32)
		kvs := ikvs.Interfaces
		kvs = append(kvs,
			"raddr", r.RemoteAddr,
			"method", r.Method,
			"path", r.URL.Path,
			"start", start.Unix(),
			"cost", cost.String(),
			"code", code,
		)

		if action != "" {
			kvs = append(kvs, "action", action)
		}

		if reqid := defaults.GetRequestID(ctx, r); reqid != "" {
			kvs = append(kvs, "reqid", reqid)
		}

		if collect != nil {
			var clean func()
			kvs, clean = collect(kvs)
			if clean != nil {
				defer clean()
			}
		}

		if err != nil {
			kvs = append(kvs, "err", err)
			if se, ok := err.(interface{ Stacks() []string }); ok {
				kvs = append(kvs, "stacks", se.Stacks())
			}
		}

		slog.Info("log http request", kvs...)
		ikvs.Interfaces = kvs
		pools.PutInterfaces(ipool, ikvs)
	})
}
