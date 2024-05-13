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

package server

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/xgfone/go-defaults"
)

func TestStartServer(t *testing.T) {
	defaults.ExitFunc.Set(func(int) {})

	func() {
		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)

		go func() {
			Start("1.2.3.4:123", nil)
			cancel()
		}()

		<-ctx.Done()
		if time.Since(start) >= time.Second {
			t.Error("expect to fail to start http server, but got success")
		}
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(204)
	})

	server := New("127.0.0.1:8800", handler)
	go Serve(server)
	time.Sleep(time.Millisecond * 100)

	resp, err := http.Get("http://127.0.0.1:8800")
	if err != nil {
		t.Error(err)
	} else {
		resp.Body.Close()
		if resp.StatusCode != 204 {
			t.Errorf("expect status code %d, but got %d", 204, resp.StatusCode)
		}
	}

	Stop(server)
	time.Sleep(time.Millisecond * 100)
}
