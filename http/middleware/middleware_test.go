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

package middleware

import (
	"net/http"
	"reflect"
	"slices"
	"testing"
)

func TestMiddlewares(t *testing.T) {
	ms := Middlewares(nil)
	ms = ms.Clone().Append(MiddlewareFunc(func(next http.Handler) http.Handler { return next }))
	if handler := ms.Handler(nil); handler != nil {
		t.Error("unexpect a http handler, but got one")
	}
}

func TestNamedPriorityMiddleware(t *testing.T) {
	ms := Middlewares{
		New("m1", 1, func(next http.Handler) http.Handler { return next }),
		New("m4", 3, func(next http.Handler) http.Handler { return next }),
		New("m2", 2, func(next http.Handler) http.Handler { return next }),
		New("m3", 3, func(next http.Handler) http.Handler { return next }),
	}
	if h := ms[0].Handler(nil); h != nil {
		t.Errorf("expect handler nil, but got %T", h)
	}

	type (
		priority interface{ Priority() int }
		namer    interface{ Name() string }
	)

	slices.SortFunc(ms, func(a, b Middleware) int {
		return a.(priority).Priority() - b.(priority).Priority()
	})

	names := make([]string, len(ms))
	for i, m := range ms {
		names[i] = m.(namer).Name()
	}

	expects := []string{"m1", "m2", "m4", "m3"}
	if !reflect.DeepEqual(names, expects) {
		t.Errorf("expect %v, but got %v", expects, names)
	}
}
