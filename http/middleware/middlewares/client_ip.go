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

package middlewares

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/xgfone/go-apiserver/http/middleware"
	"github.com/xgfone/go-apiserver/nets"
)

// ClientIP returns a new http handler middleware, which is used to check
// whehter the client ip is legal or allowed. If not, use the given handler
// to handle the response.
//
// If handler is nil, use the default handler, which returns 403.
func ClientIP(prioirty int, handler http.Handler, ipOrCidrs ...string) (middleware.Middleware, error) {
	if len(ipOrCidrs) == 0 {
		return nil, errors.New("MiddlewareClientIP: no ips or cidrs")
	}

	checkers, err := nets.NewIPCheckers(ipOrCidrs...)
	if err != nil {
		return nil, err
	}

	if handler == nil {
		handler = http.HandlerFunc(handleClientIP)
	}

	return middleware.NewMiddleware("client_ip", prioirty, func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			ip, _ := nets.SplitHostPort(r.RemoteAddr)
			if checkers.CheckIPString(ip) {
				h.ServeHTTP(rw, r)
			} else {
				handler.ServeHTTP(rw, r)
			}
		})
	}), nil
}

func handleClientIP(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(403)

	ip, _ := nets.SplitHostPort(r.RemoteAddr)
	fmt.Fprintf(rw, "the client from '%s' is not allowed", ip)
}
