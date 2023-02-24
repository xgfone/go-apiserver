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

package upstream

import (
	"context"
	"net/http"

	"github.com/xgfone/go-apiserver/http/header"
	"github.com/xgfone/go-apiserver/http/reqresp"
)

// GetRequestID is used to get the unique request session id.
//
// For the default implementation, it only detects req
// and supports the types or interfaces:
//
//	*reqresp.Context
//	*http.Request
//	interface{ GetHTTPRequest() *http.Request }
//	interface{ GetRequestID() string }
//	interface{ RequestID() string }
//
// For http.Request, it will returns the header "X-Request-Id".
//
// Return "" instead if not found.
var GetRequestID func(ctx context.Context, req interface{}) string = getRequestID

func getRequestID(ctx context.Context, req interface{}) string {
	switch r := req.(type) {
	case *reqresp.Context:
		return r.Request.Header.Get(header.HeaderXRequestID)

	case *http.Request:
		return r.Header.Get(header.HeaderXRequestID)

	case interface{ GetHTTPRequest() *http.Request }:
		return r.GetHTTPRequest().Header.Get(header.HeaderXRequestID)

	case interface{ GetRequestID() string }:
		return r.GetRequestID()

	case interface{ RequestID() string }:
		return r.RequestID()

	default:
		return ""
	}
}
