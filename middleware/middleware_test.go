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

package middleware

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/xgfone/go-apiserver/internal/test"
)

func testFuncMiddleware(buf *bytes.Buffer, name string, prio int) Middleware {
	return NewMiddleware(name, prio, func(h interface{}) interface{} {
		return func() {
			fmt.Fprintf(buf, "middleware '%s' before\n", name)
			h.(func())()
			fmt.Fprintf(buf, "middleware '%s' after\n", name)
		}
	})
}

func TestMiddlewaresSort(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	mws := Middlewares{
		testFuncMiddleware(buf, "mw2", 2),
		testFuncMiddleware(buf, "mw1", 1),
	}

	sort.Stable(mws)
	handler := mws.Handler(func() { buf.WriteString("handler\n") })
	handler.(func())()

	test.CheckStrings(t, "MiddlewareSort", strings.Split(buf.String(), "\n"), []string{
		"middleware 'mw1' before",
		"middleware 'mw2' before",
		"handler",
		"middleware 'mw2' after",
		"middleware 'mw1' after",
		"",
	})
}

func TestMiddlewaresClone(t *testing.T) {
	ms1 := Middlewares{testFuncMiddleware(nil, "m1", 1), testFuncMiddleware(nil, "m2", 2)}
	ms2 := ms1.Clone(testFuncMiddleware(nil, "m3", 3), testFuncMiddleware(nil, "m4", 4))

	if len(ms2) != 4 {
		t.Errorf("expect %d middlewares, but got %d", 4, len(ms2))
	} else {
		for i, mw := range ms2 {
			if i == 0 && mw.Name() != "m1" {
				t.Errorf("%d: expect middleware '%s', but got '%s'", i, "m1", mw.Name())
			}
			if i == 1 && mw.Name() != "m2" {
				t.Errorf("%d: expect middleware '%s', but got '%s'", i, "m2", mw.Name())
			}
			if i == 2 && mw.Name() != "m3" {
				t.Errorf("%d: expect middleware '%s', but got '%s'", i, "m3", mw.Name())
			}
			if i == 3 && mw.Name() != "m4" {
				t.Errorf("%d: expect middleware '%s', but got '%s'", i, "m4", mw.Name())
			}
		}
	}
}
