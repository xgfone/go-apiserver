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

package forwarder

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/xgfone/go-apiserver/http/handler"
)

func TestForwarder(t *testing.T) {
	const host = "127.0.0.1:8801"
	exit := make(chan struct{})
	defer func() { close(exit) }()

	go func() {
		server := &http.Server{Addr: host, Handler: handler.Handler204}
		go func() {
			<-exit
			_ = server.Shutdown(context.Background())
		}()
		_ = server.ListenAndServe()
	}()
	time.Sleep(time.Millisecond * 500)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	err := Forward(w, r, host)
	if err != nil {
		t.Fatal(err)
	}

	if w.Code != 204 {
		t.Errorf("expect statuscode %d, but got %d", 204, w.Code)
	}
}
