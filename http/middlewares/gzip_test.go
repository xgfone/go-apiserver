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
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/go-apiserver/http/header"
)

func TestGzip(t *testing.T) {
	expect := "data"

	var handler http.Handler
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(expect))
	})

	gzipMiddleware := Gzip(123, nil)
	handler = gzipMiddleware.Handler(handler).(http.Handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	req.Header.Set(header.HeaderAcceptEncoding, "gzip")
	handler.ServeHTTP(rec, req)

	r, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		t.Error(err)
	} else if s := string(data); s != expect {
		t.Errorf("expect '%s', but got '%s'", expect, s)
	}
}
