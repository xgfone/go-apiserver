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

package helper

import "io"

// JSONString represent a raw json string.
type JSONString string

// MarshalJSON implements the interface json.Marshaler.
func (s JSONString) MarshalJSON() ([]byte, error) { return []byte(s), nil }

// WriteTo implements the interface io.WriterTo, which writes the string s
// as the raw json string into w.
func (s JSONString) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, string(s))
	return int64(n), err
}

// WriteJSON is the same as WriteTo, but returns nothing.
func (s JSONString) WriteJSON(w io.Writer) { io.WriteString(w, string(s)) }
