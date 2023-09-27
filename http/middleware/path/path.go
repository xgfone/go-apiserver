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

// Package path provides a middleware to respond to the given path.
package path

import (
	"net/http"
	"strings"
)

// Repsond204 returns a middleware function to intercept the request
// matching the given path and respond status code 204 to the client,
// which is used to respond to the healthcheck in general.
//
// The trailling "/" of path will be removed.
// If path is empty, use "/" instead.
func Repsond204(path string) func(http.Handler) http.Handler {
	path = strings.TrimRight(path, "/")
	if path == "" {
		path = "/"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == path {
				w.WriteHeader(204)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}
