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

	"github.com/xgfone/go-apiserver/http/header"
)

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

	err := c.BindBody(&req)
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
