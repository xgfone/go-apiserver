// Copyright 2021~2022 xgfone
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
	"io"
	"net/http"

	"github.com/xgfone/go-apiserver/http/header"
	"github.com/xgfone/go-apiserver/result"
)

// Some HTTP errors.
var (
	ErrMissingContentType = NewError(http.StatusBadRequest).WithMsg("missing the header Content-Type")

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

// StatusCode implements the interface to return the status code.
func (e Error) StatusCode() int { return e.Code }

// CodeError implements the interface result.CodeError
// to convert itself to result.Error.
func (e Error) CodeError() result.Error {
	if e == ErrMissingContentType {
		return result.ErrMissingContentType
	}

	err := result.Error{Message: e.Error(), WrappedErr: e}
	switch e.Code {
	case http.StatusBadRequest:
		err.Code = result.CodeInvalidParams

	case http.StatusUnauthorized:
		err.Code = result.CodeUnauthorizedOperation

	case http.StatusForbidden:
		err.Code = result.CodeUnallowedOperation

	case http.StatusNotFound:
		err.Code = result.CodeInstanceNotFound

	case http.StatusMethodNotAllowed:
		err.Code = result.CodeUnallowedOperation

	case http.StatusNotAcceptable:
		err.Code = result.CodeUnsupportedOperation

	case http.StatusRequestTimeout:
		err.Code = result.CodeGatewayTimeout

	case http.StatusConflict:
		err.Code = result.CodeFailedOperation

	case http.StatusGone:
		err.Code = result.CodeInstanceUnavailable

	case http.StatusRequestEntityTooLarge:
		err.Code = result.CodeInvalidParams

	case http.StatusUnsupportedMediaType:
		err.Code = result.CodeUnsupportedMediaType

	case http.StatusTooManyRequests:
		err.Code = result.CodeRequestLimitExceeded

	case http.StatusInternalServerError:
		err.Code = result.CodeInternalServerError

	case http.StatusNotImplemented:
		err.Code = result.CodeUnsupportedOperation

	case http.StatusBadGateway:
		err.Code = result.CodeServiceUnavailable

	case http.StatusServiceUnavailable:
		err.Code = result.CodeServiceUnavailable

	case http.StatusGatewayTimeout:
		err.Code = result.CodeGatewayTimeout

	case http.StatusHTTPVersionNotSupported:
		err.Code = result.CodeUnsupportedProtocol

	default:
		err.Code = result.CodeInternalServerError
	}

	return err
}

// Error implements the interface error.
func (e Error) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return http.StatusText(e.Code)
}

// Unwrap unwraps the inner error.
func (e Error) Unwrap() error { return e.Err }

// WithCT returns a new Error with the new content type.
func (e Error) WithCT(contentType string) Error { e.CT = contentType; return e }

// WithErr returns a new Error with the new error.
func (e Error) WithErr(err error) Error { e.Err = err; return e }

// WithMsg is equal to WithErr(fmt.Errorf(msg, args...)).
func (e Error) WithMsg(msg string, args ...interface{}) Error {
	if len(args) == 0 {
		return e.WithErr(errors.New(msg))
	}
	return e.WithErr(fmt.Errorf(msg, args...))
}

// ServeHTTP implements the interface http.Handler.
func (e Error) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e.Code == 0 {
		e.Code = 200
	}

	header.SetContentType(w.Header(), e.CT)
	w.WriteHeader(e.Code)
	io.WriteString(w, e.Error())
}
