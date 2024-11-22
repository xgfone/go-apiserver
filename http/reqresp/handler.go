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
	"strconv"

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

// RespondErrorWithContextByCode parses xcode as the status code, which supports
//
//   - ""
//   - "std"
//   - number
//
// If xcode is empty, it is equal to "std".
// If xcode is a number and in [100, 599], parse it as status code.
// For other numbers or characters, but, they are equal to "std".
//
// For "std", it will guess the status code from the error.
func RespondErrorWithContextByCode(c *Context, xcode string, err error) {
	switch xcode {
	case "", "std":
		RespondErrorWithContextAndStatusCode(c, 0, err)

	case "200":
		RespondErrorWithContextAndStatusCode(c, 200, err)

	case "400":
		RespondErrorWithContextAndStatusCode(c, 400, err)

	case "500":
		RespondErrorWithContextAndStatusCode(c, 500, err)

	default:
		code, _ := strconv.ParseInt(xcode, 10, 16)
		if code >= 600 || code < 200 {
			code = 0
		}

		RespondErrorWithContextAndStatusCode(c, int(code), err)
	}
}

// If statuscode is equal to 0, guess it from the error.
func RespondErrorWithContextAndStatusCode(c *Context, statuscode int, err error) {
	if statuscode == 0 {
		responderrorstd(c, err)
	} else {
		responderror(c, statuscode, err)
	}
}

func responderror(c *Context, statuscode int, err error) {
	switch err.(type) {
	case codeint.Error, json.Marshaler:
	default:
		err = codeint.ErrInternalServerError.WithError(err)
	}

	c.JSON(statuscode, err)
}

func responderrorstd(c *Context, err error) {
	var statuscode int
	switch e := err.(type) {
	case codeint.Error, json.Marshaler:
		statuscode = getStatusCodeFromError(err)

	case StatusCoder:
		statuscode = e.StatusCode()
		err = codeint.ErrInternalServerError.WithError(err)

	default:
		statuscode = getStatusCodeFromError(err)
		err = codeint.ErrInternalServerError.WithError(err)
	}

	c.JSON(statuscode, err)
}

func getStatusCodeFromError(err error) int {
	if e, ok := err.(StatusCoder); ok {
		return e.StatusCode()
	}
	return 500
}
