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

package middlewares

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/internal/pools"
	"github.com/xgfone/go-defaults"
)

// Logger is used to wrap a http handler and return a new http handler,
// which logs the http request.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if !slog.Default().Enabled(ctx, slog.LevelInfo) {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		next.ServeHTTP(w, r)
		cost := time.Since(start)

		c := reqresp.GetContext(w, r)

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
			"start", start.Unix(),
			"cost", cost.String(),
			"code", code,
		)

		if c != nil && c.Action != "" {
			kvs = append(kvs, "action", c.Action)
		}

		if reqid := defaults.GetRequestID(ctx, r); reqid != "" {
			kvs = append(kvs, "reqid", reqid)
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