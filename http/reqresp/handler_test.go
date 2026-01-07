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

package reqresp_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-toolkit/codeint"
	"github.com/xgfone/go-toolkit/result"
)

func TestHandlerWithError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)

	c := reqresp.AcquireContext()
	resetContextResponse := func(w http.ResponseWriter) {
		c.Reset()
		c.Request = req.WithContext(reqresp.SetContext(req.Context(), c))
		c.ResponseWriter = reqresp.AcquireResponseWriter(w)
	}

	rec := httptest.NewRecorder()
	resetContextResponse(rec)
	reqresp.Handler(func(c *reqresp.Context) {
		c.ResponseWriter.WriteHeader(204)
	}).ServeHTTP(c.ResponseWriter, c.Request)
	if rec.Code != 204 {
		t.Errorf("expect status code %d, but got %d", 204, rec.Code)
	}

	rec = httptest.NewRecorder()
	resetContextResponse(rec)
	reqresp.Handler(func(c *reqresp.Context) {}).ServeHTTP(c.ResponseWriter, c.Request)
	if rec.Code != 200 {
		t.Errorf("expect status code %d, but got %d", 200, rec.Code)
	}

	rec = httptest.NewRecorder()
	resetContextResponse(rec)
	reqresp.Handler(func(c *reqresp.Context) {
		c.Err = codeint.ErrBadRequest
	}).ServeHTTP(c.ResponseWriter, c.Request)
	if rec.Code != 400 {
		t.Errorf("expect status code %d, but got %d", 400, rec.Code)
	}

	rec = httptest.NewRecorder()
	resetContextResponse(rec)
	reqresp.HandlerWithError(func(c *reqresp.Context) error {
		return nil
	}).ServeHTTP(c.ResponseWriter, c.Request)
	if rec.Code != 200 {
		t.Errorf("expect status code %d, but got %d", 200, rec.Code)
	}

	rec = httptest.NewRecorder()
	resetContextResponse(rec)
	reqresp.HandlerWithError(func(c *reqresp.Context) error {
		c.NoContent(204)
		return nil
	}).ServeHTTP(c.ResponseWriter, c.Request)
	if rec.Code != 204 {
		t.Errorf("expect status code %d, but got %d", 204, rec.Code)
	}

	rec = httptest.NewRecorder()
	resetContextResponse(rec)
	reqresp.HandlerWithError(func(c *reqresp.Context) error {
		c.NoContent(204)
		return errors.New("test")
	}).ServeHTTP(c.ResponseWriter, c.Request)
	if rec.Code != 204 {
		t.Errorf("expect status code %d, but got %d", 204, rec.Code)
	}

	rec = httptest.NewRecorder()
	resetContextResponse(rec)
	reqresp.HandlerWithError(func(c *reqresp.Context) error {
		return codeint.ErrBadRequest
	}).ServeHTTP(c.ResponseWriter, c.Request)
	if rec.Code != 400 {
		t.Errorf("expect status code %d, but got %d", 400, rec.Code)
	}

	_respond := result.GetRespondFunc()
	defer func() { result.SetRespondFunc(_respond) }()
	result.SetRespondFunc(func(responder any, _ result.Response) {
		responder.(*reqresp.Context).ResponseWriter.WriteHeader(204)
	})

	rec = httptest.NewRecorder()
	resetContextResponse(rec)
	reqresp.HandlerWithError(func(c *reqresp.Context) error {
		return codeint.ErrBadRequest
	}).ServeHTTP(c.ResponseWriter, c.Request)
	if rec.Code != 204 {
		t.Errorf("expect status code %d, but got %d", 204, rec.Code)
	}
}

func TestRespondErrorWithContextByCode(t *testing.T) {
	c := reqresp.AcquireContext()
	defer reqresp.ReleaseContext(c)

	rec := httptest.NewRecorder()
	c.ResponseWriter = reqresp.AcquireResponseWriter(rec)
	reqresp.RespondErrorWithContextByCode(c, "", codeint.ErrInternalServerError.WithReason("reason"))

	expect := `{"Code":500,"Message":"Internal Server Error"}`
	if s := strings.TrimSpace(rec.Body.String()); s != expect {
		t.Errorf("expect response body '%s', but got '%s'", expect, s)
	}
}
