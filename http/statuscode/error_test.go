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

package statuscode

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestError(t *testing.T) {
	err404 := NewError(404)
	if msg := err404.Error(); msg != http.StatusText(404) {
		t.Errorf("expect empty string, but got '%s'", msg)
	}

	if err := errors.Unwrap(err404); err != nil {
		t.Errorf("expect an error nil, but got '%v'", err)
	}

	rec := httptest.NewRecorder()
	err404.ServeHTTP(rec, nil)
	if rec.Code != 404 {
		t.Errorf("expect status code %d, but got %d", 404, rec.Code)
	} else if body := rec.Body.String(); body != "" {
		t.Errorf("expect empty body, but got '%s'", body)
	}

	err := err404.WithMessage("test")
	if msg := err.Error(); msg != "test" {
		t.Errorf("expect error '%s', but got '%s'", "test", msg)
	}

	err = err404.WithMessage("%s", "test")
	if msg := err.Error(); msg != "test" {
		t.Errorf("expect error '%s', but got '%s'", "test", msg)
	}

	if errors.Unwrap(err) == nil {
		t.Error("expect an error, but got nil")
	}

	if code := err.StatusCode(); code != 404 {
		t.Errorf("expect status code %d, but got %d", 404, code)
	}

	rec = httptest.NewRecorder()
	err.ServeHTTP(rec, nil)
	if rec.Code != 404 {
		t.Errorf("expect status code %d, but got %d", 404, rec.Code)
	} else if body := rec.Body.String(); body != "test" {
		t.Errorf("expect body '%s', but got '%s'", "test", body)
	}
}
