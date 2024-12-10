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

// Package requestid provides a request id middleware
// based on the request header "X-Request-Id".
package requestid

import (
	"net/http"

	"github.com/xgfone/go-apiserver/http/header"
	"github.com/xgfone/go-toolkit/random"
)

// Generate is used to generate a request id for the http request.
//
// Default: a random string with 24 characters
var Generate func(*http.Request) string = generate

func generate(*http.Request) string {
	return random.String(24, random.AlphaNumCharset)
}

// RequestId is a http middleware to set the request header "X-Request-Id"
// if not set.
func RequestId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(header.HeaderXRequestID) == "" {
			r.Header.Set(header.HeaderXRequestID, Generate(r))
		}
		next.ServeHTTP(w, r)
	})
}
