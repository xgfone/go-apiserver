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

// Package herrors provides some http server errors.
package herrors

import (
	"errors"
	"fmt"
	"net/http"
)

// Some non-HTTP errors
var (
	ErrMissingContentType  = errors.New("missing the header 'Content-Type'")
	ErrInvalidRedirectCode = errors.New("invalid redirect status code")
)

// Some HTTP errors.
var (
	ErrBadRequest                    = NewError(http.StatusBadRequest)
	ErrUnauthorized                  = NewError(http.StatusUnauthorized)
	ErrForbidden                     = NewError(http.StatusForbidden)
	ErrNotFound                      = NewError(http.StatusNotFound)
	ErrMethodNotAllowed              = NewError(http.StatusMethodNotAllowed)
	ErrStatusNotAcceptable           = NewError(http.StatusNotAcceptable)
	ErrRequestTimeout                = NewError(http.StatusRequestTimeout)
	ErrStatusConflict                = NewError(http.StatusConflict)
	ErrStatusGone                    = NewError(http.StatusGone)
	ErrStatusRequestEntityTooLarge   = NewError(http.StatusRequestEntityTooLarge)
	ErrUnsupportedMediaType          = NewError(http.StatusUnsupportedMediaType)
	ErrTooManyRequests               = NewError(http.StatusTooManyRequests)
	ErrInternalServerError           = NewError(http.StatusInternalServerError)
	ErrStatusNotImplemented          = NewError(http.StatusNotImplemented)
	ErrBadGateway                    = NewError(http.StatusBadGateway)
	ErrServiceUnavailable            = NewError(http.StatusServiceUnavailable)
	ErrStatusGatewayTimeout          = NewError(http.StatusGatewayTimeout)
	ErrStatusHTTPVersionNotSupported = NewError(http.StatusHTTPVersionNotSupported)
)

// Error represents a server error.
type Error struct {
	Code int
	Err  error
	CT   string // Content-Type
}

// NewError returns a new Error.
func NewError(code int) Error { return Error{Code: code} }

// Error implements the interface error.
func (e Error) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return http.StatusText(e.Code)
}

// Unwrap unwraps the inner error.
func (e Error) Unwrap() error { return e.Err }

// NewCT returns a new Error with the new ContentType ct.
func (e Error) NewCT(ct string) Error { e.CT = ct; return e }

// New returns a new Error with the new error.
func (e Error) New(err error) Error { e.Err = err; return e }

// Newf is equal to New(fmt.Errorf(msg, args...)).
func (e Error) Newf(msg string, args ...interface{}) Error {
	if len(args) == 0 {
		return e.New(errors.New(msg))
	}
	return e.New(fmt.Errorf(msg, args...))
}
