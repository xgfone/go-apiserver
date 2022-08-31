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

package rawjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func ExampleRawBytes() {
	s := RawBytes(` {"k1": "v1", "k2": 123, "k3": true} `)
	bs, _ := s.MarshalJSON()
	fmt.Println(string(bs))

	buf := bytes.NewBuffer(nil)
	s.WriteTo(buf)
	fmt.Println(buf.String())

	buf.Reset()
	s.WriteJSON(buf)
	fmt.Println(buf.String())

	data, _ := json.Marshal(s)
	fmt.Println(string(data))

	var result struct {
		K1 string `json:"k1"`
		K2 int64  `json:"k2"`
		K3 bool   `json:"k3"`
	}
	json.Unmarshal(data, &result)
	fmt.Printf("%+v", result)

	// Output:
	// {"k1": "v1", "k2": 123, "k3": true}
	// {"k1":"v1","k2":123,"k3":true}
	// {"k1":"v1","k2":123,"k3":true}
	// {"k1":"v1","k2":123,"k3":true}
	// {K1:v1 K2:123 K3:true}
}

func TestJSONBytes(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	_, err := RawBytes("  ").WriteTo(buf)
	if err != nil {
		t.Error(err)
	} else if s := buf.String(); s != `""` {
		t.Errorf(`expect json string %s, but got %s`, `""`, s)
	}
}
