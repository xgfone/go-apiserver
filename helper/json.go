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

package helper

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"sync"
)

// DecodeJSONFromString decodes the json raw string s into dst.
func DecodeJSONFromString(dst interface{}, s string) error {
	return json.NewDecoder(strings.NewReader(s)).Decode(dst)
}

// DecodeJSONFromBytes decodes the json raw bytes s into dst.
func DecodeJSONFromBytes(dst interface{}, b []byte) error {
	return json.NewDecoder(bytes.NewReader(b)).Decode(dst)
}

// EncodeJSON encodes the value by json into w.
//
// NOTICE: it does not escape the problematic HTML characters.
func EncodeJSON(w io.Writer, value interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(value)
}

// EncodeJSONBytes encodes the value by json to bytes.
//
// NOTICE: it does not escape the problematic HTML characters.
func EncodeJSONBytes(value interface{}) (data []byte, err error) {
	buf := bytes.NewBuffer(make([]byte, 0, 512))
	if err = EncodeJSON(buf, value); err == nil {
		data = buf.Bytes()
		if _len := len(data); _len > 0 && data[_len-1] == '\n' {
			data = data[:_len-1]
		}
	}
	return
}

// EncodeJSONString encodes the value by json to string.
//
// NOTICE: it does not escape the problematic HTML characters.
func EncodeJSONString(value interface{}) (data string, err error) {
	buf := getbuffer()
	if err = EncodeJSON(buf, value); err == nil {
		data = buf.String()
		if _len := len(data); _len > 0 && data[_len-1] == '\n' {
			data = data[:_len-1]
		}
	}
	putbuffer(buf)
	return
}

var bufpool = &sync.Pool{New: func() any {
	return bytes.NewBuffer(make([]byte, 0, 512))
}}

func getbuffer() *bytes.Buffer  { return bufpool.Get().(*bytes.Buffer) }
func putbuffer(b *bytes.Buffer) { b.Reset(); bufpool.Put(b) }
