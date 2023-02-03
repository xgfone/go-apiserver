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

package validator

import (
	"testing"
	"time"
)

type intComparer int

func (i intComparer) Compare(other interface{}) int {
	o := other.(intComparer)
	if i < o {
		return -1
	} else if i > o {
		return 1
	}
	return 0
}

func TestStructFieldLess(t *testing.T) {
	var v struct {
		Field1 int
		Field2 time.Time
		Field3 intComparer
	}
	v.Field1 = 123
	v.Field2 = time.Unix(123, 0)
	v.Field3 = 123

	validator1 := StructFieldLess("Field1")
	if err := validator1.Validate(v, 100); err != nil {
		t.Errorf("expect nil, but got an error: %v", err)
	}
	if err := validator1.Validate(v, 123); err == nil {
		t.Errorf("expect an error, but got nil")
	}
	if err := validator1.Validate(v, 200); err == nil {
		t.Errorf("expect an error, but got nil")
	}

	validator2 := StructFieldLess("Field2")
	if err := validator2.Validate(v, time.Unix(100, 0)); err != nil {
		t.Errorf("expect nil, but got an error: %v", err)
	}
	if err := validator2.Validate(v, time.Unix(123, 0)); err == nil {
		t.Errorf("expect an error, but got nil")
	}
	if err := validator2.Validate(v, time.Unix(200, 0)); err == nil {
		t.Errorf("expect an error, but got nil")
	}

	validator3 := StructFieldLess("Field3")
	if err := validator3.Validate(v, intComparer(100)); err != nil {
		t.Errorf("expect nil, but got an error: %v", err)
	}
	if err := validator3.Validate(v, intComparer(123)); err == nil {
		t.Errorf("expect an error, but got nil")
	}
	if err := validator3.Validate(v, intComparer(200)); err == nil {
		t.Errorf("expect an error, but got nil")
	}
}

func TestStructFieldGreater(t *testing.T) {
	var v struct {
		Field1 int
		Field2 time.Time
		Field3 intComparer
	}
	v.Field1 = 123
	v.Field2 = time.Unix(123, 0)
	v.Field3 = 123

	validator1 := StructFieldGreater("Field1")
	if err := validator1.Validate(v, 200); err != nil {
		t.Errorf("expect nil, but got an error: %v", err)
	}
	if err := validator1.Validate(v, 123); err == nil {
		t.Errorf("expect an error, but got nil")
	}
	if err := validator1.Validate(v, 100); err == nil {
		t.Errorf("expect an error, but got nil")
	}

	validator2 := StructFieldGreater("Field2")
	if err := validator2.Validate(v, time.Unix(200, 0)); err != nil {
		t.Errorf("expect nil, but got an error: %v", err)
	}
	if err := validator2.Validate(v, time.Unix(123, 0)); err == nil {
		t.Errorf("expect an error, but got nil")
	}
	if err := validator2.Validate(v, time.Unix(100, 0)); err == nil {
		t.Errorf("expect an error, but got nil")
	}

	validator3 := StructFieldGreater("Field3")
	if err := validator3.Validate(v, intComparer(200)); err != nil {
		t.Errorf("expect nil, but got an error: %v", err)
	}
	if err := validator3.Validate(v, intComparer(123)); err == nil {
		t.Errorf("expect an error, but got nil")
	}
	if err := validator3.Validate(v, intComparer(100)); err == nil {
		t.Errorf("expect an error, but got nil")
	}
}

func TestStructFieldLessEqual(t *testing.T) {
	var v struct {
		Field1 int
		Field2 time.Time
		Field3 intComparer
	}
	v.Field1 = 123
	v.Field2 = time.Unix(123, 0)
	v.Field3 = 123

	validator1 := StructFieldLessEqual("Field1")
	if err := validator1.Validate(v, 100); err != nil {
		t.Errorf("expect nil, but got an error: %v", err)
	}
	if err := validator1.Validate(v, 123); err != nil {
		t.Errorf("expect nil, but got an error: %v", err)
	}
	if err := validator1.Validate(v, 200); err == nil {
		t.Errorf("expect an error, but got nil")
	}

	validator2 := StructFieldLessEqual("Field2")
	if err := validator2.Validate(v, time.Unix(100, 0)); err != nil {
		t.Errorf("expect nil, but got an error: %v", err)
	}
	if err := validator2.Validate(v, time.Unix(123, 0)); err != nil {
		t.Errorf("expect nil, but got an error: %v", err)
	}
	if err := validator2.Validate(v, time.Unix(200, 0)); err == nil {
		t.Errorf("expect an error, but got nil")
	}

	validator3 := StructFieldLessEqual("Field3")
	if err := validator3.Validate(v, intComparer(100)); err != nil {
		t.Errorf("expect nil, but got an error: %v", err)
	}
	if err := validator3.Validate(v, intComparer(123)); err != nil {
		t.Errorf("expect nil, but got an error: %v", err)
	}
	if err := validator3.Validate(v, intComparer(200)); err == nil {
		t.Errorf("expect an error, but got nil")
	}
}

func TestStructFieldGreaterEqual(t *testing.T) {
	var v struct {
		Field1 int
		Field2 time.Time
		Field3 intComparer
	}
	v.Field1 = 123
	v.Field2 = time.Unix(123, 0)
	v.Field3 = 123

	validator1 := StructFieldGreaterEqual("Field1")
	if err := validator1.Validate(v, 200); err != nil {
		t.Errorf("expect nil, but got an error: %v", err)
	}
	if err := validator1.Validate(v, 123); err != nil {
		t.Errorf("expect nil, but got an error: %v", err)
	}
	if err := validator1.Validate(v, 100); err == nil {
		t.Errorf("expect an error, but got nil")
	}

	validator2 := StructFieldGreaterEqual("Field2")
	if err := validator2.Validate(v, time.Unix(200, 0)); err != nil {
		t.Errorf("expect nil, but got an error: %v", err)
	}
	if err := validator2.Validate(v, time.Unix(123, 0)); err != nil {
		t.Errorf("expect nil, but got an error: %v", err)
	}
	if err := validator2.Validate(v, time.Unix(100, 0)); err == nil {
		t.Errorf("expect an error, but got nil")
	}

	validator3 := StructFieldGreaterEqual("Field3")
	if err := validator3.Validate(v, intComparer(200)); err != nil {
		t.Errorf("expect nil, but got an error: %v", err)
	}
	if err := validator3.Validate(v, intComparer(123)); err != nil {
		t.Errorf("expect nil, but got an error: %v", err)
	}
	if err := validator3.Validate(v, intComparer(100)); err == nil {
		t.Errorf("expect an error, but got nil")
	}
}
