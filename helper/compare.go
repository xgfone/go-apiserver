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
	"math/big"
	"reflect"
	"time"
)

// Comparer is used to compare with the other.
type Comparer interface {
	// The returned value is
	//   -1 when it is less than other.
	//    0 when they are equal.
	//    1 when it is greater than other.
	Compare(other interface{}) int
}

// Compare compares the size of the left and right values, and returns
//   -1 when left is less than right, that's, left < right.
//    0 when left is equal to right, that's, left == right.
//    1 when left is greater than right, that's, left > right.
//
// Support the types as following:
//   int, int8, int16, int32, int64
//   uint, uint8, uint16, uint32, uint64
//   float32, float64
//   time.Time
//   Comparer
// And the customized types based on the base integer and float types.
//
// Notice: Both of left and right are the same type.
func Compare(left, right interface{}) int {
	switch v1 := left.(type) {
	case int:
		return compareInt(int64(v1), int64(right.(int)))
	case int8:
		return compareInt(int64(v1), int64(right.(int8)))
	case int16:
		return compareInt(int64(v1), int64(right.(int16)))
	case int32:
		return compareInt(int64(v1), int64(right.(int32)))
	case int64:
		return compareInt(int64(v1), int64(right.(int64)))

	case uint:
		return compareUint(uint64(v1), uint64(right.(uint)))
	case uint8:
		return compareUint(uint64(v1), uint64(right.(uint8)))
	case uint16:
		return compareUint(uint64(v1), uint64(right.(uint16)))
	case uint32:
		return compareUint(uint64(v1), uint64(right.(uint32)))
	case uint64:
		return compareUint(uint64(v1), uint64(right.(uint64)))

	case float32:
		return big.NewFloat(float64(v1)).Cmp(big.NewFloat(float64(right.(float32))))
	case float64:
		return big.NewFloat(v1).Cmp(big.NewFloat(right.(float64)))

	case time.Time:
		if v2 := right.(time.Time); v1.Before(v2) {
			return -1
		} else if v1.After(v2) {
			return 1
		}
		return 0

	case Comparer:
		return v1.Compare(right)

	default:
		switch vf := reflect.ValueOf(v1); vf.Kind() {
		case reflect.Float32, reflect.Float64:
			return big.NewFloat(vf.Float()).Cmp(big.NewFloat(reflect.ValueOf(right).Float()))

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return compareInt(vf.Int(), reflect.ValueOf(right).Int())

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return compareUint(vf.Uint(), reflect.ValueOf(right).Uint())

		default:
			panic(fmt.Errorf("Compare: unknown type %T", left))
		}
	}
}

// LT reports whether left is less than right, which is eqaul to
//   Compare(left, right) == -1
func LT(left, right interface{}) bool { return Compare(left, right) == -1 }

// LE reports whether left is less than or equal to right, which is eqaul to
//   Compare(left, right) != 1
func LE(left, right interface{}) bool { return Compare(left, right) != 1 }

// GT reports whether left is greater than right, which is eqaul to
//   Compare(left, right) == 1
func GT(left, right interface{}) bool { return Compare(left, right) == 1 }

// GE reports whether left is greater than or equal to right, which is eqaul to
//   Compare(left, right) != -1
func GE(left, right interface{}) bool { return Compare(left, right) != -1 }

func compareInt(left, right int64) int {
	if left < right {
		return -1
	} else if left > right {
		return 1
	}
	return 0
}

func compareUint(left, right uint64) int {
	if left < right {
		return -1
	} else if left > right {
		return 1
	}
	return 0
}
