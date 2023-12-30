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

import "net/http"

// DefaultHandler is handle the response when not respond to client.
//
// Default: nil
var DefaultHandler func(c *Context)

func handleContextResult(c *Context) {
	if c.ResponseWriter.WroteHeader() {
		return
	}

	switch {
	case DefaultHandler != nil:
		DefaultHandler(c)

	case c.Err == nil:
		c.WriteHeader(200)

	case !respondError(c, c.Err):
		c.Text(500, c.Err.Error())
	}
}

func respondError(c *Context, err error) (ok bool) {
	switch e := err.(type) {
	case http.Handler:
		e.ServeHTTP(c.ResponseWriter, c.Request)
		ok = true

	case interface{ Unwrap() error }:
		if err = e.Unwrap(); err != nil {
			ok = respondError(c, err)
		}

	case interface{ Unwrap() []error }:
		for _, err := range e.Unwrap() {
			if ok = respondError(c, err); ok {
				break
			}
		}
	}
	return
}

// Handler is a handler based on Context to handle the http request.
type Handler func(c *Context)

// ServeHTTP implements the interface http.Handler.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := GetContext(r.Context())
	if c == nil {
		c = AcquireContext()
		defer ReleaseContext(c)

		c.Request = r.WithContext(SetContext(r.Context(), c))
		c.ResponseWriter = AcquireResponseWriter(w)
		defer ReleaseResponseWriter(c.ResponseWriter)
	}

	h(c)
	handleContextResult(c)
}

// HandlerWithError is a handler to handle the http request with the error.
type HandlerWithError func(c *Context) error

// ServeHTTP implements the interface http.Handler.
func (h HandlerWithError) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := GetContext(r.Context())
	if c == nil {
		c = AcquireContext()
		defer ReleaseContext(c)

		c.Request = r.WithContext(SetContext(r.Context(), c))
		c.ResponseWriter = AcquireResponseWriter(w)
		defer ReleaseResponseWriter(c.ResponseWriter)
	}

	c.AppendError(h(c))
	handleContextResult(c)
}
