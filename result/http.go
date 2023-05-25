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

package result

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/xgfone/go-apiserver/http/header"
)

var _ http.Handler = Error{}

// ServeHTTP implements the interface http.Handler.
func (e Error) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	header.SetContentType(w.Header(), header.MIMEApplicationJSONCharsetUTF8)
	w.WriteHeader(ToHttpStatusCode(e.Code))
	json.NewEncoder(w).Encode(e)
}

// ToHttpStatusCode converts the code to the http status code.
var ToHttpStatusCode func(code string) (httpStatusCode int) = toHttpStatusCode

func toHttpStatusCode(code string) int {
	switch code {
	case "":
		return http.StatusOK // 200

	case CodeBadRequest:
		return http.StatusBadRequest // 400

	case CodeAuthFailure:
		return http.StatusUnauthorized // 401

	case CodeUnallowedUnauthorized:
		return http.StatusForbidden // 403

	case CodeNotFound:
		return http.StatusNotFound // 404

	case CodeUnallowedInconsistent:
		return http.StatusConflict // 409

	case CodeBadRequestUnsupportedMediaType:
		return http.StatusUnsupportedMediaType // 415

	case CodeUnallowedExceedLimit:
		return http.StatusTooManyRequests // 429

	case CodeInternalServerErrorBadGateway:
		return http.StatusBadGateway // 502

	case CodeInternalServerErrorUnavailable:
		return http.StatusServiceUnavailable // 503

	case CodeInternalServerErrorTimeout:
		return http.StatusGatewayTimeout // 504

	default:
		switch {
		case strings.HasPrefix(code, "BadRequest."):
			return http.StatusBadRequest // 400
		case strings.HasPrefix(code, "AuthFailure."):
			return http.StatusUnauthorized // 401
		case strings.HasPrefix(code, "NotFound."):
			return http.StatusNotFound // 404
		case strings.HasPrefix(code, "Unallowed.ExceedLimit."):
			return http.StatusTooManyRequests // 429

		default:
			return http.StatusInternalServerError // 500
		}
	}
}
