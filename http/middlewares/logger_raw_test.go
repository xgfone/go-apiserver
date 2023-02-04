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

package middlewares

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestRawString(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	s := rawString(` {"k1": "v1", "k2": 123, "k3": true} `)
	if err := json.NewEncoder(buf).Encode(s); err != nil {
		t.Fatal(err)
	}

	expect := `{"k1":"v1","k2":123,"k3":true}` + "\n"
	if result := buf.String(); result != expect {
		t.Errorf("expect '%s', but got '%s'", expect, result)
	}

	var result struct {
		K1 string `json:"k1"`
		K2 int64  `json:"k2"`
		K3 bool   `json:"k3"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.K1 != "v1" || result.K2 != 123 || !result.K3 {
		t.Errorf("unexpect result %+v", result)
	}
}
