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

import "github.com/xgfone/go-toolkit/codeint"

// Pre-define some errors with the status code.
var (
	ErrBadRequest           = codeint.ErrBadRequest           // 400
	ErrUnauthorized         = codeint.ErrUnauthorized         // 401
	ErrForbidden            = codeint.ErrForbidden            // 403
	ErrNotFound             = codeint.ErrNotFound             // 404
	ErrConflict             = codeint.ErrConflict             // 409
	ErrUnsupportedMediaType = codeint.ErrUnsupportedMediaType // 415
	ErrTooManyRequests      = codeint.ErrTooManyRequests      // 429
	ErrInternalServerError  = codeint.ErrInternalServerError  // 500
	ErrBadGateway           = codeint.ErrBadGateway           // 502
	ErrServiceUnavailable   = codeint.ErrServiceUnavailable   // 503
	ErrGatewayTimeout       = codeint.ErrGatewayTimeout       // 504

	ErrMissingContentType = codeint.ErrMissingContentType
)
