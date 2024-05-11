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

// Package path provides a middleware to respond to the given path.
package path

import (
	"net/http"
	"strings"
)

// Repsond204 is equal to Repsond(204, path).
func Repsond204(path string) func(http.Handler) http.Handler {
	return Repsond(204, path)
}

// Repsond is a http middleware function to intercept the request
// matching the given path and respond status code to the client,
// which is used to respond to the healthcheck in general.
//
// The trailling "/" of path will be removed.
// If path is empty, use "/" instead.
func Repsond(code int, path string) func(http.Handler) http.Handler {
	path = strings.TrimRight(path, "/")
	if path == "" {
		path = "/"
	}

	pathlen := len(path)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == path || (len(r.URL.Path) == pathlen+1 && r.URL.Path[pathlen] == '/') {
				w.WriteHeader(204)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}
