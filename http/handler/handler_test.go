// Copyright 2021~2023 xgfone
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

package handler

import (
	"net/http/httptest"
	"testing"
)

func TestSwitchHandler(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expect a panic, but got not")
			}
		}()
		NewSwitchHandler(nil)
	}()

	h := NewSwitchHandler(Handler200)

	rec := httptest.NewRecorder()
	h.Set(Handler204)
	h.Get().ServeHTTP(rec, nil)
	if rec.Code != 204 {
		t.Errorf("expect status code %d, but got %d", 204, rec.Code)
	}

	rec = httptest.NewRecorder()
	h.Swap(Handler400).ServeHTTP(rec, nil)
	if rec.Code != 204 {
		t.Errorf("expect status code %d, but got %d", 204, rec.Code)
	}

	rec = httptest.NewRecorder()
	h.Get().ServeHTTP(rec, nil)
	if rec.Code != 400 {
		t.Errorf("expect status code %d, but got %d", 400, rec.Code)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, nil)
	if rec.Code != 400 {
		t.Errorf("expect status code %d, but got %d", 400, rec.Code)
	}
}
