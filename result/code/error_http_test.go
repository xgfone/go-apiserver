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

package code

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/result"
)

func TestErrorHttp(t *testing.T) {
	respond := result.Respond
	defer func() { result.Respond = respond }()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	result.Respond = func(responder any, resp result.Response) {
		if r, ok := responder.(*httptest.ResponseRecorder); !ok {
			t.Errorf("expect *httptest.ResponseRecorder, but got %T", responder)
		} else {
			r.WriteHeader(resp.StatusCode())
		}
	}
	ErrBadRequest.ServeHTTPWithCode(rec, req, 500)
	if rec.Code != 500 {
		t.Errorf("expect status code %d, but got %d", 500, rec.Code)
	}

	rec = httptest.NewRecorder()
	c := reqresp.AcquireContext()
	c.ResponseWriter = reqresp.AcquireResponseWriter(rec)
	req = req.WithContext(reqresp.SetContext(req.Context(), c))
	result.Respond = func(responder any, resp result.Response) {
		if r, ok := responder.(*reqresp.Context); !ok {
			t.Errorf("expect *reqresp.Context, but got %T", responder)
		} else {
			r.WriteHeader(resp.StatusCode())
		}
	}
	ErrBadRequest.ServeHTTPWithCode(rec, req, 501)
	if rec.Code != 501 {
		t.Errorf("expect status code %d, but got %d", 501, rec.Code)
	}
}
