// Copyright 2021 xgfone
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

package reqresp

import (
	"net/http"

	"github.com/xgfone/go-apiserver/result"
)

// Handler is a handler based on Context to handle the http request.
type Handler func(c *Context)

// ServeHTTP implements the interface http.Handler.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	runhandler(w, r, h)
}

// HandlerWithError is a handler to handle the http request with the error.
type HandlerWithError func(c *Context) error

// ServeHTTP implements the interface http.Handler.
func (h HandlerWithError) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	runhandler(w, r, func(c *Context) { c.AppendError(h(c)) })
}

func runhandler(w http.ResponseWriter, r *http.Request, f Handler) {
	c := GetContext(r.Context())
	if c == nil {
		c = AcquireContext()
		defer ReleaseContext(c)

		c.Request = r.WithContext(SetContext(r.Context(), c))
		c.ResponseWriter = AcquireResponseWriter(w)
		defer ReleaseResponseWriter(c.ResponseWriter)
	}

	if f(c); !c.ResponseWriter.WroteHeader() {
		result.Err(c.Err).Respond(c)
	}
}
