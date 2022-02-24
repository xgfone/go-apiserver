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

package reqresp

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/xgfone/go-apiserver/helper"
	"github.com/xgfone/go-apiserver/http/header"
)

func BenchmarkRequestSetDataGetData(b *testing.B) {
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	req = SetReqData(req, "key", "value")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req = SetReqData(req, "key", "value")
		if value, ok := GetReqData(req, "key").(string); !ok || value != "value" {
			panic("invalid value")
		}
	}
}

func TestRequestParams(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	req = SetReqData(req, "key1", "value1")
	req = SetReqDatas(req, map[string]interface{}{"key2": "value2"})
	if value, _ := GetReqData(req, "key1").(string); value != "value1" {
		t.Errorf("expect '%s', but got '%s'", "value1", value)
	}

	datas := GetReqDatas(req)
	if len(datas) != 2 {
		t.Errorf("expect %d parameters, but got %d", 2, len(datas))
	} else {
		expects := []string{"key1", "key2"}
		for key := range datas {
			if !helper.InStrings(key, expects) {
				t.Errorf("unexpect key '%s'", key)
			}
		}
	}

	DefaultContextAllocator.Release(GetContext(req))
}

func TestContextBinder(t *testing.T) {
	var req struct {
		Int    int    `default:"222"`
		Uint   int    `default:"333"`
		String string `default:"abc"`
	}
	req.Int = 111

	c := DefaultContextAllocator.Acquire()
	body := bytes.NewBufferString(`{"Uint":444}`)
	c.Request, _ = http.NewRequest("GET", "http://localhost", body)
	c.Request.Header.Set(header.HeaderContentType, header.MIMEApplicationJSON)

	err := c.Bind(&req)
	if err != nil {
		t.Error(err)
	}

	if req.Int != 111 {
		t.Errorf("expect Int is equal to %d, but got %d", 111, req.Int)
	}
	if req.Uint != 444 {
		t.Errorf("expect Uint is equal to %d, but got %d", 444, req.Uint)
	}
	if req.String != "abc" {
		t.Errorf("expect String is equal to '%s', but '%s'", "abc", req.String)
	}
}
