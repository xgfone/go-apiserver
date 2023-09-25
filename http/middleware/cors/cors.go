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

// Package cors provides a CORS middleware.
package cors

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/xgfone/go-apiserver/http/middleware"
)

// DefaultAllowMethods is the default allowed methods,
// which is set as the value of the http header "Access-Control-Allow-Methods".
var DefaultAllowMethods = []string{
	http.MethodHead, http.MethodGet,
	http.MethodPost, http.MethodPut,
	http.MethodPatch, http.MethodDelete,
}

// Config is used to configure the CORS middleware.
type Config struct {
	// AllowOrigin defines a list of origins that may access the resource.
	//
	// Optional. Default: []string{"*"}.
	AllowOrigins []string `json:"allowOrigins" yaml:"allowOrigins"`

	// AllowHeaders indicates a list of request headers used in response to
	// a preflight request to indicate which HTTP headers can be used when
	// making the actual request. This is in response to a preflight request.
	//
	// Optional. Default: []string{}.
	AllowHeaders []string `json:"allowHeaders" yaml:"allowHeaders"`

	// AllowMethods indicates methods allowed when accessing the resource.
	// This is used in response to a preflight request.
	//
	// Optional. Default: DefaultAllowMethods.
	AllowMethods []string `json:"allowMethods" yaml:"allowMethods"`

	// ExposeHeaders indicates a server whitelist headers that browsers are
	// allowed to access. This is in response to a preflight request.
	//
	// Optional. Default: []string{}.
	ExposeHeaders []string `json:"exposeHeaders" yaml:"exposeHeaders"`

	// AllowCredentials indicates whether or not the response to the request
	// can be exposed when the credentials flag is true. When used as part of
	// a response to a preflight request, this indicates whether or not the
	// actual request can be made using credentials.
	//
	// Optional. Default: false.
	AllowCredentials bool `json:"allowCredentials" yaml:"allowCredentials"`

	// MaxAge indicates how long (in seconds) the results of a preflight request
	// can be cached.
	//
	// Optional. Default: 0.
	MaxAge int `json:"maxAge" yaml:"maxAge"`
}

// CORS returns a new middleware named "cors", which implements HTTP CORS protocol.
// see https://fetch.spec.whatwg.org/#http-cors-protocol.
func CORS(config Config) middleware.MiddlewareFunc {
	if len(config.AllowOrigins) == 0 {
		config.AllowOrigins = []string{"*"}
	}
	if len(config.AllowMethods) == 0 {
		config.AllowMethods = DefaultAllowMethods
	}

	allowMethods := strings.Join(config.AllowMethods, ",")
	allowHeaders := strings.Join(config.AllowHeaders, ",")
	exposeHeaders := strings.Join(config.ExposeHeaders, ",")
	maxAge := fmt.Sprintf("%d", config.MaxAge)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check whether the origin is allowed or not.
			var allowOrigin string
			origin := r.Header.Get("Origin")

		LOOP:
			for _, o := range config.AllowOrigins {
				switch {
				case o == "*":
					if config.AllowCredentials {
						allowOrigin = origin
					} else {
						allowOrigin = o
					}
					break LOOP

				case o == origin:
					allowOrigin = o
					break LOOP

				default:
					if matchSubdomain(origin, o) {
						allowOrigin = origin
						break LOOP
					}
				}
			}

			if len(allowOrigin) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			respHeader := w.Header()
			respHeader.Add("Vary", "Origin")
			respHeader.Set("Access-Control-Allow-Origin", allowOrigin)
			if config.AllowCredentials {
				respHeader.Set("Access-Control-Allow-Credentials", "true")
			}

			if r.Method != http.MethodOptions {
				// Simple request
				if exposeHeaders != "" {
					respHeader.Set("Access-Control-Expose-Headers", exposeHeaders)
				}
				next.ServeHTTP(w, r)
				return
			}

			// Preflight request
			respHeader.Add("Vary", "Access-Control-Request-Method")
			respHeader.Add("Vary", "Access-Control-Request-Headers")
			respHeader.Set("Access-Control-Allow-Methods", allowMethods)

			if allowHeaders != "" {
				respHeader.Set("Access-Control-Allow-Headers", allowHeaders)
			} else if h := r.Header.Get("Access-Control-Request-Headers"); h != "" {
				respHeader.Set("Access-Control-Allow-Headers", h)
			}

			if config.MaxAge > 0 {
				respHeader.Set("Access-Control-Max-Age", maxAge)
			}

			w.WriteHeader(http.StatusNoContent)
		})
	}

}

func matchScheme(domain, pattern string) bool {
	didx := strings.Index(domain, ":")
	pidx := strings.Index(pattern, ":")
	return didx != -1 && pidx != -1 && domain[:didx] == pattern[:pidx]
}

// matchSubdomain compares authority with wildcard
func matchSubdomain(domain, pattern string) bool {
	if !matchScheme(domain, pattern) {
		return false
	}
	didx := strings.Index(domain, "://")
	pidx := strings.Index(pattern, "://")
	if didx == -1 || pidx == -1 {
		return false
	}
	domAuth := domain[didx+3:]
	// to avoid long loop by invalid long domain
	if len(domAuth) > 253 {
		return false
	}
	patAuth := pattern[pidx+3:]

	domComp := strings.Split(domAuth, ".")
	patComp := strings.Split(patAuth, ".")
	for i := len(domComp)/2 - 1; i >= 0; i-- {
		opp := len(domComp) - 1 - i
		domComp[i], domComp[opp] = domComp[opp], domComp[i]
	}
	for i := len(patComp)/2 - 1; i >= 0; i-- {
		opp := len(patComp) - 1 - i
		patComp[i], patComp[opp] = patComp[opp], patComp[i]
	}

	for i, v := range domComp {
		if len(patComp) <= i {
			return false
		}
		p := patComp[i]
		if p == "*" {
			return true
		}
		if p != v {
			return false
		}
	}
	return false
}
