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
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestError(t *testing.T) {
	err404 := NewError(404)
	if msg := err404.Error(); msg != "code=404, msg=Not Found" {
		t.Errorf("expect '%s', but got '%s'", "code=404, msg=Not Found", msg)
	}
	if msg := err404.WithMessagef("test").Error(); msg != "code=404, msg=test" {
		t.Errorf("expect '%s', but got '%s'", "404: test", msg)
	}

	if err := errors.Unwrap(err404); err != nil {
		t.Errorf("expect an error nil, but got '%v'", err)
	}

	if code := err404.StatusCode(); code != 404 {
		t.Errorf("expect status code %d, but got %d", 404, code)
	}

	rec := httptest.NewRecorder()
	expect := `{"Code":404,"Message":"test"}`
	err404.WithMessagef("%s", "test").ServeHTTP(rec, nil)
	if rec.Code != 404 {
		t.Errorf("expect status code %d, but got %d", 404, rec.Code)
	} else if body := strings.TrimSpace(rec.Body.String()); body != expect {
		t.Errorf("expect '%s', but got '%s'", expect, body)
	}

	if reason := err404.GetReason(); reason != `Not Found` {
		t.Errorf("expect reason '%s', but got '%s'", `Not Found`, reason)
	}
}
