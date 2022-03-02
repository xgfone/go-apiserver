// Copyright 2022 xgfone
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

package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/go-apiserver/http/handler"
)

func TestClientIP(t *testing.T) {
	clientIP, err := ClientIP(0, nil, "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	handler := clientIP.Handler(handler.Handler400).(http.Handler)

	rec1 := httptest.NewRecorder()
	req.RemoteAddr = "10.1.2.3:12345"
	handler.ServeHTTP(rec1, req)
	if rec1.Code != 400 {
		t.Errorf("expect status code '%d', but got '%d'", 400, rec1.Code)
	}

	rec2 := httptest.NewRecorder()
	req.RemoteAddr = "172.16.1.2:12345"
	handler.ServeHTTP(rec2, req)
	if rec2.Code != 400 {
		t.Errorf("expect status code '%d', but got '%d'", 400, rec2.Code)
	}

	rec3 := httptest.NewRecorder()
	req.RemoteAddr = "192.168.1.2:12345"
	handler.ServeHTTP(rec3, req)
	if rec3.Code != 400 {
		t.Errorf("expect status code '%d', but got '%d'", 400, rec3.Code)
	}

	rec4 := httptest.NewRecorder()
	req.RemoteAddr = "1.2.3.4:12345"
	handler.ServeHTTP(rec4, req)
	if rec4.Code != 403 {
		t.Errorf("expect status code '%d', but got '%d'", 403, rec4.Code)
	}
}
