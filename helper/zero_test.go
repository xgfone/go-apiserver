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

import "testing"

func TestIsZero(t *testing.T) {
	if !IsZero(nil) {
		t.Error("expect ZERO, but got not")
	}
	if !IsZero(false) {
		t.Error("expect ZERO, but got not")
	}
	if !IsZero("") {
		t.Error("expect ZERO, but got not")
	}
	if !IsZero(int(0)) {
		t.Error("expect ZERO, but got not")
	}
	if !IsZero(int8(0)) {
		t.Error("expect ZERO, but got not")
	}
	if !IsZero(int16(0)) {
		t.Error("expect ZERO, but got not")
	}
	if !IsZero(int32(0)) {
		t.Error("expect ZERO, but got not")
	}
	if !IsZero(int64(0)) {
		t.Error("expect ZERO, but got not")
	}
	if !IsZero(uint(0)) {
		t.Error("expect ZERO, but got not")
	}
	if !IsZero(uint8(0)) {
		t.Error("expect ZERO, but got not")
	}
	if !IsZero(uint16(0)) {
		t.Error("expect ZERO, but got not")
	}
	if !IsZero(uint32(0)) {
		t.Error("expect ZERO, but got not")
	}
	if !IsZero(uint64(0)) {
		t.Error("expect ZERO, but got not")
	}
	if !IsZero(uintptr(0)) {
		t.Error("expect ZERO, but got not")
	}
	if !IsZero(float32(0)) {
		t.Error("expect ZERO, but got not")
	}
	if !IsZero(float64(0)) {
		t.Error("expect ZERO, but got not")
	}
	if !IsZero([]byte(nil)) {
		t.Error("expect ZERO, but got not")
	}
	if !IsZero(zero(123)) {
		t.Error("expect ZERO, but got not")
	}
	if !IsZero(struct{}{}) {
		t.Error("expect ZERO, but got not")
	}
}

type zero int

func (z zero) IsZero() bool { return true }
