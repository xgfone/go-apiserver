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
	"strings"
	"testing"

	"github.com/xgfone/go-apiserver/helper"
)

func TestRecover(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test")
	})

	var stacks []string
	DefaultPanicHandler = func(w http.ResponseWriter, r *http.Request, recover interface{}) {
		for _, stack := range helper.GetCallStack(5) {
			if strings.HasPrefix(stack, "testing/") || strings.HasPrefix(stack, "net/http/") {
				continue
			}
			stacks = append(stacks, stack)
		}
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://127.0.0.1", nil)
	Recover(0).Handler(handler).(http.Handler).ServeHTTP(rec, req)

	expects := []string{
		"github.com/xgfone/go-apiserver/http/middlewares/recover_test.go:func1:28",
		"github.com/xgfone/go-apiserver/http/middlewares/recover.go:1:51",
		"github.com/xgfone/go-apiserver/http/middlewares/recover_test.go:TestRecover:43",
	}

	if len(expects) != len(stacks) {
		t.Fatalf("expect %d line, but got %d: %v", len(expects), len(stacks), stacks)
	}

	for i, line := range expects {
		if line != stacks[i] {
			t.Errorf("%d: expect '%s', but got '%s'", i, line, stacks[i])
		}
	}
}
