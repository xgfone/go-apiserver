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

package handler

import (
	"net/http"
	"testing"

	"github.com/xgfone/go-apiserver/helper"
)

type testHandler struct{ name string }

func (h testHandler) String() string                               { return h.name }
func (h testHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}

func TestSwitchHandlerUnwrap(t *testing.T) {
	handler := testHandler{name: "test"}
	sh := NewSwitchHandler(handler)

	if h, _ := helper.Unwrap[http.Handler](sh); h != handler {
		t.Errorf("expect '%v', but got '%v'", handler, h)
	}
}
