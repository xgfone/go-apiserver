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

// Package statuscode provides an error based on the status code.
package statuscode

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Pre-define some errors with the status code.
var (
	ErrMissingContentType = NewError(http.StatusBadRequest).WithMessage("missing the header Content-Type")

	ErrBadRequest           = NewError(http.StatusBadRequest)           // 400
	ErrUnauthorized         = NewError(http.StatusUnauthorized)         // 401
	ErrForbidden            = NewError(http.StatusForbidden)            // 403
	ErrNotFound             = NewError(http.StatusNotFound)             // 404
	ErrConflict             = NewError(http.StatusConflict)             // 409
	ErrUnsupportedMediaType = NewError(http.StatusUnsupportedMediaType) // 415
	ErrTooManyRequests      = NewError(http.StatusTooManyRequests)      // 429
	ErrInternalServerError  = NewError(http.StatusInternalServerError)  // 500
	ErrBadGateway           = NewError(http.StatusBadGateway)           // 502
	ErrServiceUnavailable   = NewError(http.StatusServiceUnavailable)   // 503
	ErrGatewayTimeout       = NewError(http.StatusGatewayTimeout)       // 504
)

var (
	_ error        = Error{}
	_ http.Handler = Error{}
)

// Error represents an error with the status code.
type Error struct {
	Code int
	Err  error
}

// NewError returns a new Error with the status code and without error.
func NewError(statusCode int) Error { return Error{Code: statusCode} }

// StatusCode returns the error status code.
func (e Error) StatusCode() int { return e.Code }

// Unwrap returns the wrapped error.
func (e Error) Unwrap() error { return e.Err }

// Error returns the error message.
func (e Error) Error() string {
	if e.Err == nil {
		return http.StatusText(e.Code)
	}
	return e.Err.Error()
}

// WithError returns a new Error with the new error.
func (e Error) WithError(err error) Error {
	e.Err = err
	return e
}

// WithMessage is a convenient method that convert the format message
// to an error and set it, then return the new error.
func (e Error) WithMessage(msg string, args ...interface{}) Error {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	return e.WithError(errors.New(msg))
}

// ServeHTTP implements the interface http.Handler.
func (e Error) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e.Err == nil {
		w.WriteHeader(e.Code)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(e.Code)
	_, _ = io.WriteString(w, e.Err.Error())
}
