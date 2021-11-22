// Copyright 2021 xgfone
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

package http

import (
	"net/http"
	"testing"
)

func TestRequestParams(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	req = SetParam(req, "key1", "value1")
	req = SetParams(req, map[string]string{"key2": "value2"})

	if value, _ := GetParam(req, "key1"); value != "value1" {
		t.Errorf("expect '%s', but got '%s'", "value1", value)
	}

	params := GetParams(req)
	if len(params) != 2 {
		t.Errorf("expect %d parameters, but got %d", 2, len(params))
	} else {
		for key := range params {
			if key != "key1" && key != "key2" {
				t.Errorf("unexpect key '%s'", key)
			}
		}
	}
}

func BenchmarkRequestParams(b *testing.B) {
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	req = SetParam(req, "key", "value")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req = SetParam(req, "key", "value")
		GetParam(req, "key")
	}
}
