// Copyright 2024 xgfone
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

package codeint

import (
	"net/http"

	"github.com/xgfone/go-apiserver/http/handler"
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

// StatusCode returns the http status code.
//
// If Status is not equal to 0, return it.
// Or, return Code if it is in [100, 599].
// Or, return 500.
func (e Error) StatusCode() int {
	if e.Status != 0 {
		return e.Status
	}
	if 100 <= e.Code && e.Code < 600 {
		return e.Code
	}
	return 500
}

// ServeHTTP implements the interface http.Handler.
func (e Error) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_ = handler.JSON(w, e.StatusCode(), e)
}
