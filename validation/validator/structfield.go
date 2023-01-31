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
	"fmt"
	"time"

	"github.com/xgfone/go-apiserver/helper"
)

// StructFieldLess returns a validator to check whether the current value
// is less than other struct field named by fieldName.
//
// Notice: it only supports the types as following,
//
//	uint, uint8, uint16, uint32, uint64
//	int, int8, int16, int32, int64
//	float32, float64
//	time.Time
//
// The validator rule is "ltf(fieldName)".
func StructFieldLess(fieldName string) Validator {
	if fieldName == "" {
		panic("the field name must not be empty")
	}

	rule := fmt.Sprintf("ltf(%s)", fieldName)
	return NewValidator(rule, func(ctx, value interface{}) error {
		if ctx == nil {
			return fmt.Errorf("the context is nil")
		}

		vf, ok := helper.GetStructFieldByName(ctx, fieldName)
		if !ok {
			panic(fmt.Errorf("the context is not a struct or does not contain the field named '%s'", fieldName))
		}

		if helper.IsPointer(vf) {
			if vf.IsNil() {
				return fmt.Errorf("the field named '%s' is nil", fieldName)
			}
			vf = vf.Elem()
		}

		return isLess(helper.Indirect(value), vf.Interface(), fieldName)
	})
}

// StructFieldLessEqual returns a validator to check whether the current value
// is less than or equal to other struct field named by fieldName.
//
// Notice: it only supports the types as following,
//
//	uint, uint8, uint16, uint32, uint64
//	int, int8, int16, int32, int64
//	float32, float64
//	time.Time
//
// The validator rule is "lef(fieldName)".
func StructFieldLessEqual(fieldName string) Validator {
	if fieldName == "" {
		panic("the field name must not be empty")
	}

	rule := fmt.Sprintf("lef(%s)", fieldName)
	return NewValidator(rule, func(ctx, value interface{}) error {
		if ctx == nil {
			return fmt.Errorf("the context is nil")
		}

		vf, ok := helper.GetStructFieldByName(ctx, fieldName)
		if !ok {
			panic(fmt.Errorf("the context is not a struct or does not contain the field named '%s'", fieldName))
		}

		if helper.IsPointer(vf) {
			if vf.IsNil() {
				return fmt.Errorf("the field named '%s' is nil", fieldName)
			}
			vf = vf.Elem()
		}

		return isLessEqual(helper.Indirect(value), vf.Interface(), fieldName)
	})
}

// StructFieldGreater returns a validator to check whether the current value
// is greater than other struct field named by fieldName.
//
// Notice: it only supports the types as following,
//
//	uint, uint8, uint16, uint32, uint64
//	int, int8, int16, int32, int64
//	float32, float64
//	time.Time
//
// The validator rule is "gtf(fieldName)".
func StructFieldGreater(fieldName string) Validator {
	if fieldName == "" {
		panic("the field name must not be empty")
	}

	rule := fmt.Sprintf("gtf(%s)", fieldName)
	return NewValidator(rule, func(ctx, value interface{}) error {
		if ctx == nil {
			return fmt.Errorf("the context is nil")
		}

		vf, ok := helper.GetStructFieldByName(ctx, fieldName)
		if !ok {
			panic(fmt.Errorf("the context is not a struct or does not contain the field named '%s'", fieldName))
		}

		if helper.IsPointer(vf) {
			if vf.IsNil() {
				return fmt.Errorf("the field named '%s' is nil", fieldName)
			}
			vf = vf.Elem()
		}

		return isGreater(helper.Indirect(value), vf.Interface(), fieldName)
	})
}

// StructFieldGreaterEqual returns a validator to check whether the current value
// is greater than or equal to other struct field named by fieldName.
//
// Notice: it only supports the types as following,
//
//	uint, uint8, uint16, uint32, uint64
//	int, int8, int16, int32, int64
//	float32, float64
//	time.Time
//
// The validator rule is "gef(fieldName)".
func StructFieldGreaterEqual(fieldName string) Validator {
	if fieldName == "" {
		panic("the field name must not be empty")
	}

	rule := fmt.Sprintf("gef(%s)", fieldName)
	return NewValidator(rule, func(ctx, value interface{}) error {
		if ctx == nil {
			return fmt.Errorf("the context is nil")
		}

		vf, ok := helper.GetStructFieldByName(ctx, fieldName)
		if !ok {
			panic(fmt.Errorf("the context is not a struct or does not contain the field named '%s'", fieldName))
		}

		if helper.IsPointer(vf) {
			if vf.IsNil() {
				return fmt.Errorf("the field named '%s' is nil", fieldName)
			}
			vf = vf.Elem()
		}

		return isGreaterEqual(helper.Indirect(value), vf.Interface(), fieldName)
	})
}

func isLess(left, right interface{}, rightName string) error {
	switch v1 := left.(type) {
	case int:
		if v2, ok := right.(int); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 < v2) {
			return fmt.Errorf("the value is not less than the field named '%s'", rightName)
		}

	case int8:
		if v2, ok := right.(int8); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 < v2) {
			return fmt.Errorf("the value is not less than the field named '%s'", rightName)
		}

	case int16:
		if v2, ok := right.(int16); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 < v2) {
			return fmt.Errorf("the value is not less than the field named '%s'", rightName)
		}

	case int32:
		if v2, ok := right.(int32); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 < v2) {
			return fmt.Errorf("the value is not less than the field named '%s'", rightName)
		}

	case int64:
		if v2, ok := right.(int64); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 < v2) {
			return fmt.Errorf("the value is not less than the field named '%s'", rightName)
		}

	case uint:
		if v2, ok := right.(uint); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 < v2) {
			return fmt.Errorf("the value is not less than the field named '%s'", rightName)
		}

	case uint8:
		if v2, ok := right.(uint8); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 < v2) {
			return fmt.Errorf("the value is not less than the field named '%s'", rightName)
		}

	case uint16:
		if v2, ok := right.(uint16); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 < v2) {
			return fmt.Errorf("the value is not less than the field named '%s'", rightName)
		}

	case uint32:
		if v2, ok := right.(uint32); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 < v2) {
			return fmt.Errorf("the value is not less than the field named '%s'", rightName)
		}

	case uint64:
		if v2, ok := right.(uint64); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 < v2) {
			return fmt.Errorf("the value is not less than the field named '%s'", rightName)
		}

	case float32:
		if v2, ok := right.(float32); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 < v2) {
			return fmt.Errorf("the value is not less than the field named '%s'", rightName)
		}

	case float64:
		if v2, ok := right.(float64); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 < v2) {
			return fmt.Errorf("the value is not less than the field named '%s'", rightName)
		}

	case time.Time:
		if v2, ok := right.(time.Time); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !v1.Before(v2) {
			return fmt.Errorf("the value is not less than the field named '%s'", rightName)
		}

	case helper.Comparer:
		switch v := v1.Compare(right); v {
		case -1:
		case 0, 1:
			return fmt.Errorf("the value is not less than the field named '%s'", rightName)
		default:
			panic(fmt.Errorf("Compare returns an unknown result %d", v))
		}

	default:
		panic(fmt.Errorf("not support the type %T", left))
	}

	return nil
}

func isLessEqual(left, right interface{}, rightName string) error {
	switch v1 := left.(type) {
	case int:
		if v2, ok := right.(int); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 <= v2) {
			return fmt.Errorf("the value is not less than or equal to the field named '%s'", rightName)
		}

	case int8:
		if v2, ok := right.(int8); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 <= v2) {
			return fmt.Errorf("the value is not less than or equal to the field named '%s'", rightName)
		}

	case int16:
		if v2, ok := right.(int16); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 <= v2) {
			return fmt.Errorf("the value is not less than or equal to the field named '%s'", rightName)
		}

	case int32:
		if v2, ok := right.(int32); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 <= v2) {
			return fmt.Errorf("the value is not less than or equal to the field named '%s'", rightName)
		}

	case int64:
		if v2, ok := right.(int64); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 <= v2) {
			return fmt.Errorf("the value is not less than or equal to the field named '%s'", rightName)
		}

	case uint:
		if v2, ok := right.(uint); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 <= v2) {
			return fmt.Errorf("the value is not less than or equal to the field named '%s'", rightName)
		}

	case uint8:
		if v2, ok := right.(uint8); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 <= v2) {
			return fmt.Errorf("the value is not less than or equal to the field named '%s'", rightName)
		}

	case uint16:
		if v2, ok := right.(uint16); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 <= v2) {
			return fmt.Errorf("the value is not less than or equal to the field named '%s'", rightName)
		}

	case uint32:
		if v2, ok := right.(uint32); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 <= v2) {
			return fmt.Errorf("the value is not less than or equal to the field named '%s'", rightName)
		}

	case uint64:
		if v2, ok := right.(uint64); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 <= v2) {
			return fmt.Errorf("the value is not less than or equal to the field named '%s'", rightName)
		}

	case float32:
		if v2, ok := right.(float32); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 <= v2) {
			return fmt.Errorf("the value is not less than or equal to the field named '%s'", rightName)
		}

	case float64:
		if v2, ok := right.(float64); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 <= v2) {
			return fmt.Errorf("the value is not less than or equal to the field named '%s'", rightName)
		}

	case time.Time:
		if v2, ok := right.(time.Time); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1.Before(v2) || v1.Equal(v2)) {
			return fmt.Errorf("the value is not less than the field named '%s'", rightName)
		}

	case helper.Comparer:
		switch v := v1.Compare(right); v {
		case -1, 0:
		case 1:
			return fmt.Errorf("the value is not less than or equal to the field named '%s'", rightName)
		default:
			panic(fmt.Errorf("Compare returns an unknown result %d", v))
		}

	default:
		panic(fmt.Errorf("not support the type %T", left))
	}

	return nil
}

func isGreater(left, right interface{}, rightName string) error {
	switch v1 := left.(type) {
	case int:
		if v2, ok := right.(int); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 > v2) {
			return fmt.Errorf("the value is not greater than the field named '%s'", rightName)
		}

	case int8:
		if v2, ok := right.(int8); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 > v2) {
			return fmt.Errorf("the value is not greater than the field named '%s'", rightName)
		}

	case int16:
		if v2, ok := right.(int16); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 > v2) {
			return fmt.Errorf("the value is not greater than the field named '%s'", rightName)
		}

	case int32:
		if v2, ok := right.(int32); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 > v2) {
			return fmt.Errorf("the value is not greater than the field named '%s'", rightName)
		}

	case int64:
		if v2, ok := right.(int64); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 > v2) {
			return fmt.Errorf("the value is not greater than the field named '%s'", rightName)
		}

	case uint:
		if v2, ok := right.(uint); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 > v2) {
			return fmt.Errorf("the value is not greater than the field named '%s'", rightName)
		}

	case uint8:
		if v2, ok := right.(uint8); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 > v2) {
			return fmt.Errorf("the value is not greater than the field named '%s'", rightName)
		}

	case uint16:
		if v2, ok := right.(uint16); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 > v2) {
			return fmt.Errorf("the value is not greater than the field named '%s'", rightName)
		}

	case uint32:
		if v2, ok := right.(uint32); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 > v2) {
			return fmt.Errorf("the value is not greater than the field named '%s'", rightName)
		}

	case uint64:
		if v2, ok := right.(uint64); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 > v2) {
			return fmt.Errorf("the value is not greater than the field named '%s'", rightName)
		}

	case float32:
		if v2, ok := right.(float32); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 > v2) {
			return fmt.Errorf("the value is not greater than the field named '%s'", rightName)
		}

	case float64:
		if v2, ok := right.(float64); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 > v2) {
			return fmt.Errorf("the value is not greater than the field named '%s'", rightName)
		}

	case time.Time:
		if v2, ok := right.(time.Time); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !v1.After(v2) {
			return fmt.Errorf("the value is not less than the field named '%s'", rightName)
		}

	case helper.Comparer:
		switch v := v1.Compare(right); v {
		case 1:
		case 0, -1:
			return fmt.Errorf("the value is not greater than the field named '%s'", rightName)
		default:
			panic(fmt.Errorf("Compare returns an unknown result %d", v))
		}

	default:
		panic(fmt.Errorf("not support the type %T", left))
	}

	return nil
}

func isGreaterEqual(left, right interface{}, rightName string) error {
	switch v1 := left.(type) {
	case int:
		if v2, ok := right.(int); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 >= v2) {
			return fmt.Errorf("the value is not greater than or equal to the field named '%s'", rightName)
		}

	case int8:
		if v2, ok := right.(int8); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 >= v2) {
			return fmt.Errorf("the value is not greater than or equal to the field named '%s'", rightName)
		}

	case int16:
		if v2, ok := right.(int16); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 >= v2) {
			return fmt.Errorf("the value is not greater than or equal to the field named '%s'", rightName)
		}

	case int32:
		if v2, ok := right.(int32); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 >= v2) {
			return fmt.Errorf("the value is not greater than or equal to the field named '%s'", rightName)
		}

	case int64:
		if v2, ok := right.(int64); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 >= v2) {
			return fmt.Errorf("the value is not greater than or equal to the field named '%s'", rightName)
		}

	case uint:
		if v2, ok := right.(uint); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 >= v2) {
			return fmt.Errorf("the value is not greater than or equal to the field named '%s'", rightName)
		}

	case uint8:
		if v2, ok := right.(uint8); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 >= v2) {
			return fmt.Errorf("the value is not greater than or equal to the field named '%s'", rightName)
		}

	case uint16:
		if v2, ok := right.(uint16); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 >= v2) {
			return fmt.Errorf("the value is not greater than or equal to the field named '%s'", rightName)
		}

	case uint32:
		if v2, ok := right.(uint32); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 >= v2) {
			return fmt.Errorf("the value is not greater than or equal to the field named '%s'", rightName)
		}

	case uint64:
		if v2, ok := right.(uint64); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 >= v2) {
			return fmt.Errorf("the value is not greater than or equal to the field named '%s'", rightName)
		}

	case float32:
		if v2, ok := right.(float32); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 >= v2) {
			return fmt.Errorf("the value is not greater than or equal to the field named '%s'", rightName)
		}

	case float64:
		if v2, ok := right.(float64); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1 >= v2) {
			return fmt.Errorf("the value is not greater than or equal to the field named '%s'", rightName)
		}

	case time.Time:
		if v2, ok := right.(time.Time); !ok {
			panic(fmt.Errorf("the type is not consistent with the field named '%s'", rightName))
		} else if !(v1.After(v2) || v1.Equal(v2)) {
			return fmt.Errorf("the value is not less than the field named '%s'", rightName)
		}

	case helper.Comparer:
		switch v := v1.Compare(right); v {
		case 1, 0:
		case -1:
			return fmt.Errorf("the value is not greater than or equal to the field named '%s'", rightName)
		default:
			panic(fmt.Errorf("Compare returns an unknown result %d", v))
		}

	default:
		panic(fmt.Errorf("not support the type %T", left))
	}

	return nil
}
