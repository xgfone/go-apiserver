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

import (
	"fmt"
	"testing"
	"time"
)

func ExampleCompare() {
	// The customize base integer types
	type Int int
	type Int8 int8
	type Int16 int16
	type Int32 int32
	type Int64 int64
	type Uint uint
	type Uint8 uint8
	type Uint16 uint16
	type Uint32 uint32
	type Uint64 uint64
	type Float32 float32
	type Float64 float64

	fmt.Println(Compare(int(123), int(200)))
	fmt.Println(Compare(int8(123), int8(100)))
	fmt.Println(Compare(int16(123), int16(200)))
	fmt.Println(Compare(int32(123), int32(100)))
	fmt.Println(Compare(int64(123), int64(200)))
	fmt.Println(Compare(uint(123), uint(100)))
	fmt.Println(Compare(uint8(123), uint8(200)))
	fmt.Println(Compare(uint16(123), uint16(100)))
	fmt.Println(Compare(uint32(123), uint32(200)))
	fmt.Println(Compare(uint64(123), uint64(100)))
	fmt.Println(Compare(float32(123), float32(200)))
	fmt.Println(Compare(float64(123), float64(100)))
	fmt.Println(Compare(time.Unix(123, 0), time.Unix(200, 0)))

	fmt.Println()

	fmt.Println(Compare(Int(123), Int(100)))
	fmt.Println(Compare(Int8(123), Int8(125)))
	fmt.Println(Compare(Int16(123), Int16(100)))
	fmt.Println(Compare(Int32(123), Int32(200)))
	fmt.Println(Compare(Int64(123), Int64(100)))
	fmt.Println(Compare(Uint(123), Uint(200)))
	fmt.Println(Compare(Uint8(123), Uint8(100)))
	fmt.Println(Compare(Uint16(123), Uint16(200)))
	fmt.Println(Compare(Uint32(123), Uint32(100)))
	fmt.Println(Compare(Uint64(123), Uint64(200)))
	fmt.Println(Compare(Float32(123), Float32(100)))
	fmt.Println(Compare(Float64(123), Float64(200)))
	fmt.Println(Compare(time.Unix(123, 0), time.Unix(100, 0)))

	fmt.Println()

	fmt.Println(Compare(Int(123), Int(123)))
	fmt.Println(Compare(Int8(123), Int8(123)))
	fmt.Println(Compare(Int16(123), Int16(123)))
	fmt.Println(Compare(Int32(123), Int32(123)))
	fmt.Println(Compare(Int64(123), Int64(123)))
	fmt.Println(Compare(Uint(123), Uint(123)))
	fmt.Println(Compare(Uint8(123), Uint8(123)))
	fmt.Println(Compare(Uint16(123), Uint16(123)))
	fmt.Println(Compare(Uint32(123), Uint32(123)))
	fmt.Println(Compare(Uint64(123), Uint64(123)))
	fmt.Println(Compare(Float32(123), Float32(123)))
	fmt.Println(Compare(Float64(123), Float64(123)))
	fmt.Println(Compare(time.Unix(123, 0), time.Unix(123, 0)))

	// Output:
	// -1
	// 1
	// -1
	// 1
	// -1
	// 1
	// -1
	// 1
	// -1
	// 1
	// -1
	// 1
	// -1
	//
	// 1
	// -1
	// 1
	// -1
	// 1
	// -1
	// 1
	// -1
	// 1
	// -1
	// 1
	// -1
	// 1
	//
	// 0
	// 0
	// 0
	// 0
	// 0
	// 0
	// 0
	// 0
	// 0
	// 0
	// 0
	// 0
	// 0
}

func TestLT(t *testing.T) {
	if !LT(123, 456) {
		t.Error("expect true, bug got false")
	}

	if LT(123, 123) {
		t.Error("expect false, bug got true")
	}

	if LT(456, 123) {
		t.Error("expect false, bug got true")
	}
}

func TestLE(t *testing.T) {
	if !LE(123, 456) {
		t.Error("expect true, bug got false")
	}

	if !LE(123, 123) {
		t.Error("expect true, bug got false")
	}

	if LE(456, 123) {
		t.Error("expect false, bug got true")
	}
}

func TestGT(t *testing.T) {
	if !GT(456, 123) {
		t.Error("expect true, bug got false")
	}

	if GT(456, 456) {
		t.Error("expect false, bug got true")
	}

	if GT(123, 456) {
		t.Error("expect false, bug got true")
	}
}

func TestGE(t *testing.T) {
	if !GE(456, 123) {
		t.Error("expect true, bug got false")
	}

	if !GE(456, 456) {
		t.Error("expect true, bug got false")
	}

	if GE(123, 456) {
		t.Error("expect false, bug got true")
	}
}
