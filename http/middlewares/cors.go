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
	"fmt"
	"net/http"
	"strings"

	"github.com/xgfone/go-apiserver/http/header"
	"github.com/xgfone/go-apiserver/middleware"
)

// CORSConfig is used to configure the CORS middleware.
type CORSConfig struct {
	// AllowOrigin defines a list of origins that may access the resource.
	//
	// Optional. Default: []string{"*"}.
	AllowOrigins []string

	// AllowHeaders indicates a list of request headers used in response to
	// a preflight request to indicate which HTTP headers can be used when
	// making the actual request. This is in response to a preflight request.
	//
	// Optional. Default: []string{}.
	AllowHeaders []string

	// AllowMethods indicates methods allowed when accessing the resource.
	// This is used in response to a preflight request.
	//
	// Optional. Default: []string{"HEAD", "GET", "POST", "PUT", "PATHC", "DELETE"}.
	AllowMethods []string

	// ExposeHeaders indicates a server whitelist headers that browsers are
	// allowed to access. This is in response to a preflight request.
	//
	// Optional. Default: []string{}.
	ExposeHeaders []string

	// AllowCredentials indicates whether or not the response to the request
	// can be exposed when the credentials flag is true. When used as part of
	// a response to a preflight request, this indicates whether or not the
	// actual request can be made using credentials.
	//
	// Optional. Default: false.
	AllowCredentials bool

	// MaxAge indicates how long (in seconds) the results of a preflight request
	// can be cached.
	//
	// Optional. Default: 0.
	MaxAge int
}

// CORS returns a CORS middleware.
func CORS(priority int, config *CORSConfig) middleware.Middleware {
	var conf CORSConfig
	if config != nil {
		conf = *config
	}

	if len(conf.AllowOrigins) == 0 {
		conf.AllowOrigins = []string{"*"}
	}
	if len(conf.AllowMethods) == 0 {
		conf.AllowMethods = []string{http.MethodHead, http.MethodGet,
			http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}
	}

	allowMethods := strings.Join(conf.AllowMethods, ",")
	allowHeaders := strings.Join(conf.AllowHeaders, ",")
	exposeHeaders := strings.Join(conf.ExposeHeaders, ",")
	maxAge := fmt.Sprintf("%d", conf.MaxAge)

	return middleware.NewMiddleware("cors", priority, func(h interface{}) interface{} {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check whether the origin is allowed or not.
			var allowOrigin string
			origin := r.Header.Get(header.HeaderOrigin)
			for _, o := range conf.AllowOrigins {
				if o == "*" {
					if conf.AllowCredentials {
						allowOrigin = origin
					} else {
						allowOrigin = o
					}
				} else if o == origin {
					allowOrigin = o
					break
				}

				if matchSubdomain(origin, o) {
					allowOrigin = origin
					break
				}
			}

			if len(allowOrigin) == 0 {
				h.(http.Handler).ServeHTTP(w, r)
				return
			}

			respHeader := w.Header()
			respHeader.Add(header.HeaderVary, header.HeaderOrigin)
			respHeader.Set(header.HeaderAccessControlAllowOrigin, allowOrigin)
			if conf.AllowCredentials {
				respHeader.Set(header.HeaderAccessControlAllowCredentials, "true")
			}

			if r.Method != http.MethodOptions {
				// Simple request
				if exposeHeaders != "" {
					respHeader.Set(header.HeaderAccessControlExposeHeaders, exposeHeaders)
				}
				h.(http.Handler).ServeHTTP(w, r)

			} else {
				// Preflight request
				respHeader.Add(header.HeaderVary, header.HeaderAccessControlRequestMethod)
				respHeader.Add(header.HeaderVary, header.HeaderAccessControlRequestHeaders)
				respHeader.Set(header.HeaderAccessControlAllowMethods, allowMethods)

				if allowHeaders != "" {
					respHeader.Set(header.HeaderAccessControlAllowHeaders, allowHeaders)
				} else if h := r.Header.Get(header.HeaderAccessControlRequestHeaders); h != "" {
					respHeader.Set(header.HeaderAccessControlAllowHeaders, h)
				}

				if conf.MaxAge > 0 {
					respHeader.Set(header.HeaderAccessControlMaxAge, maxAge)
				}

				w.WriteHeader(http.StatusNoContent)
			}
		})
	})
}
