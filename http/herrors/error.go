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

// ContentType returns the Content-Type to render the error.
func (e Error) ContentType() string { return e.CT }

// CodeError implements the interface result.CodeError
// to convert itself to result.Error.
//
// If e.Err is a result.Error, return it directly.
func (e Error) CodeError() result.Error {
	if e == ErrMissingContentType {
		return result.ErrMissingContentType
	}

	if err, ok := e.Err.(result.Error); ok {
		return err
	}

	return result.Error{
		Code:       GetCodeByStatus(e.Code),
		Message:    e.Error(),
		WrappedErr: e,
	}
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

var status2code = make(map[int]string, 32)

// MappingStatusCode adds the mapping from status to code.
func MappingStatusCode(fromStatus int, toCode string) {
	status2code[fromStatus] = toCode
}

// GetCodeByStatus gets the code by the status by the mapping from status to code.
func GetCodeByStatus(status int) (code string) {
	if len(status2code) > 0 {
		if code, ok := status2code[status]; ok {
			return code
		}
	}
	return getCodeDefault(status)
}

func getCodeDefault(status int) string {
	switch status {
	case http.StatusBadRequest: // 400
		return result.CodeInvalidParams

	case http.StatusUnauthorized: // 401
		return result.CodeUnauthorizedOperation

	case http.StatusForbidden: // 403
		return result.CodeUnallowedOperation

	case http.StatusNotFound: // 404
		return result.CodeInstanceNotFound

	case http.StatusMethodNotAllowed: // 405
		return result.CodeUnallowedOperation

	case http.StatusNotAcceptable: // 406
		return result.CodeUnsupportedOperation

	case http.StatusRequestTimeout: // 408
		return result.CodeGatewayTimeout

	case http.StatusConflict: // 409
		return result.CodeFailedOperation

	case http.StatusGone: // 410
		return result.CodeInstanceUnavailable

	case http.StatusLengthRequired: // 411
		return result.CodeInvalidParams

	case http.StatusRequestEntityTooLarge: // 413
		return result.CodeInvalidParams

	case http.StatusRequestURITooLong: // 414
		return result.CodeInvalidParams

	case http.StatusUnsupportedMediaType: // 415
		return result.CodeUnsupportedMediaType

	case http.StatusExpectationFailed: // 417
		return result.CodeInternalServerError

	case http.StatusUnprocessableEntity: // 422
		return result.CodeFailedOperation

	case http.StatusTooManyRequests: // 429
		return result.CodeRequestLimitExceeded

	case http.StatusRequestHeaderFieldsTooLarge: // 431
		return result.CodeInvalidParams

	case http.StatusInternalServerError: // 500
		return result.CodeInternalServerError

	case http.StatusNotImplemented: // 501
		return result.CodeUnsupportedOperation

	case http.StatusBadGateway: // 502
		return result.CodeServiceUnavailable

	case http.StatusServiceUnavailable: // 503
		return result.CodeServiceUnavailable

	case http.StatusGatewayTimeout: // 504
		return result.CodeGatewayTimeout

	case http.StatusHTTPVersionNotSupported: // 505
		return result.CodeUnsupportedProtocol

	default:
		switch {
		case status < 400:
			return ""
		case status < 500:
			return result.CodeInvalidParams
		default:
			return result.CodeInternalServerError
		}
	}
}
