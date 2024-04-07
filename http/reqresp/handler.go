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
	"encoding/json"
	"net/http"

	"github.com/xgfone/go-apiserver/result"
	"github.com/xgfone/go-apiserver/result/codeint"
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

/// ----------------------------------------------------------------------- ///

func RespondResultResponse(c *Context, response result.Response) {
	xcode := c.Request.Header.Get("X-Response-Code")
	if xcode == "" {
		xcode = c.GetQuery("X-Response-Code")
	}

	switch xcode {
	case "", "std":
		respondstd(c, response)
	case "200":
		respond200(c, response)
	case "500":
		respond500(c, response)
	default:
		respondstd(c, response)
	}
}

func getStatusCodeFromError(err error) int {
	if e, ok := err.(StatusCoder); ok {
		return e.StatusCode()
	}
	return 500
}

func respondstd(c *Context, r result.Response) {
	switch e := r.Error.(type) {
	case nil:
		c.JSON(200, r.Data)

	case codeint.Error, json.Marshaler:
		c.JSON(getStatusCodeFromError(r.Error), r.Error)

	case StatusCoder:
		r.Error = codeint.ErrInternalServerError.WithError(r.Error)
		c.JSON(e.StatusCode(), r.Error)

	default:
		r.Error = codeint.ErrInternalServerError.WithError(r.Error)
		c.JSON(getStatusCodeFromError(r.Error), r.Error)
	}
}

func respond200(c *Context, r result.Response) {
	switch r.Error.(type) {
	case nil:
		c.JSON(200, r.Data)

	case codeint.Error, json.Marshaler:
		c.JSON(200, r.Error)

	default:
		c.JSON(200, codeint.ErrInternalServerError.WithError(r.Error))
	}
}

func respond500(c *Context, r result.Response) {
	switch r.Error.(type) {
	case nil:
		c.JSON(200, r.Data)

	case codeint.Error, json.Marshaler:
		c.JSON(500, r.Error)

	default:
		c.JSON(500, codeint.ErrInternalServerError.WithError(r.Error))
	}
}
