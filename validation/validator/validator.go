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

// Package validator provides a validator interface and some implementations.
package validator

import (
	"fmt"
	"reflect"
	"strings"
)

// ValueValidator represents the interface implemented by the value.
//
// If a value has implemented the interface, it can validate itself.
type ValueValidator interface {
	Validate(ctx interface{}) error
}

// Validator is a validator to check whether the given value is valid.
type Validator interface {
	Validate(ctx, value interface{}) error
	String() string
}

// ValidatorFunc represents a validation function.
type ValidatorFunc func(interface{}, interface{}) error

// NewValidator returns the new Validator based on the validation rule
// and function.
func NewValidator(rule string, validate ValidatorFunc) Validator {
	return validator{s: rule, f: validate}
}

type validator struct {
	s string
	f ValidatorFunc
}

func (v validator) Validate(c, i interface{}) error { return v.f(c, i) }
func (v validator) String() string                  { return v.s }

// BoolValidatorFunc converts a bool validation function to ValidatorFunc,
// which returns err if validate returns false, or nil if true.
func BoolValidatorFunc(validate func(interface{}) bool, err error) ValidatorFunc {
	if err == nil {
		panic("BoolValidatorFunc: the error must not be nil")
	}
	if validate == nil {
		panic("BoolValidatorFunc: the validation function must not be nil")
	}

	return func(_, i interface{}) error {
		if validate(i) {
			return nil
		}
		return err
	}
}

// StringBoolValidatorFunc converts a bool validation function to ValidatorFunc,
// which returns err if validate returns false, or nil if true.
func StringBoolValidatorFunc(validate func(string) bool, err error) ValidatorFunc {
	if err == nil {
		panic("StringBoolValidatorFunc: the error must not be nil")
	}
	if validate == nil {
		panic("StringBoolValidatorFunc: the validation function must not be nil")
	}

	return func(_, i interface{}) error {
		var ok bool
		switch t := i.(type) {
		case string:
			ok = validate(t)

		case fmt.Stringer:
			ok = validate(t.String())

		default:
			return fmt.Errorf("unsupported type '%T'", i)
		}

		if ok {
			return nil
		}
		return err
	}
}

func formatValidators(sep string, validators []Validator) string {
	switch len(validators) {
	case 0:
		return ""
	case 1:
		return validators[0].String()
	}

	var b strings.Builder
	b.Grow(32)

	b.WriteByte('(')
	for i, validator := range validators {
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(validator.String())
	}
	b.WriteByte(')')

	return b.String()
}

func composeValidators(name string, validators ...Validator) (Validator, string) {
	validator := And(validators...)
	desc := validator.String()
	if desc[0] == '(' {
		desc = name + desc
	} else {
		desc = fmt.Sprintf("%s(%s)", name, desc)
	}
	return validator, desc
}

// ************************************************************************* //

// AndValidator is a And validator based on a set of the validators.
type andValidator []Validator

// Validate implements the interface Validator.
func (vs andValidator) Validate(c, v interface{}) (err error) {
	for i, _len := 0, len(vs); i < _len; i++ {
		if err = vs[i].Validate(c, v); err != nil {
			return
		}
	}
	return
}

func (vs andValidator) String() string {
	return formatValidators(" && ", []Validator(vs))
}

// And returns a new And Validator.
func And(validators ...Validator) Validator {
	switch len(validators) {
	case 0:
		panic("AndValidator: no validators")
	case 1:
		return validators[0]
	}

	vs := make(andValidator, 0, len(validators))
	for _, v := range validators {
		if andv, ok := v.(andValidator); ok {
			vs = append(vs, []Validator(andv)...)
		} else {
			vs = append(vs, v)
		}
	}

	return andValidator(vs)
}

// ************************************************************************* //

// OrValidator is a OR validator based on a set of the validators.
type orValidator []Validator

// Validate implements the interface Validator.
func (vs orValidator) Validate(c, v interface{}) (err error) {
	for i, _len := 0, len(vs); i < _len; i++ {
		if err = vs[i].Validate(c, v); err == nil {
			return nil
		}
	}
	return
}

func (vs orValidator) String() string {
	return formatValidators(" || ", []Validator(vs))
}

// Or returns a new OR Validator.
func Or(validators ...Validator) Validator {
	switch len(validators) {
	case 0:
		panic("OrValidator: no validators")
	case 1:
		return validators[0]
	}

	vs := make(orValidator, 0, len(validators))
	for _, v := range validators {
		if orv, ok := v.(orValidator); ok {
			vs = append(vs, []Validator(orv)...)
		} else {
			vs = append(vs, v)
		}
	}

	return orValidator(vs)
}

// ************************************************************************* //

// Array returns a new Validator to use the given validators to check
// each element of the array or slice.
func Array(validators ...Validator) Validator {
	if len(validators) == 0 {
		panic("ArrayValidator: need at least one validator")
	}

	validator, desc := composeValidators("array", validators...)
	return NewValidator(desc, func(c, i interface{}) error {
		switch vs := i.(type) {
		case []string:
			for i, s := range vs {
				if err := validator.Validate(c, s); err != nil {
					return fmt.Errorf("%dth element is invalid: %v", i, err)
				}
			}

		default:
			vf := reflect.ValueOf(i)
			if vf.Kind() == reflect.Ptr {
				vf = vf.Elem()
			}
			switch vf.Kind() {
			case reflect.Slice, reflect.Array:
			default:
				return fmt.Errorf("expect the value is a slice or array, but got %T", i)
			}

			for i, _len := 0, vf.Len(); i < _len; i++ {
				vf.Index(i).Interface()
				if err := validator.Validate(c, vf.Index(i).Interface()); err != nil {
					return fmt.Errorf("%dth element is invalid: %v", i, err)
				}
			}
		}

		return nil
	})
}

// ************************************************************************* //

// MapK returns a new Validator to use the given validators to check
// each key of the map.
func MapK(validators ...Validator) Validator {
	if len(validators) == 0 {
		panic("MapKValidator: need at least one validator")
	}

	validator, desc := composeValidators("mapk", validators...)
	return NewValidator(desc, func(c, i interface{}) error {
		switch vs := i.(type) {
		case map[string]string:
			for key := range vs {
				if err := validator.Validate(c, key); err != nil {
					return fmt.Errorf("map key '%s' is invalid: %v", key, err)
				}
			}

		case map[string]interface{}:
			for key := range vs {
				if err := validator.Validate(c, key); err != nil {
					return fmt.Errorf("map key '%s' is invalid: %v", key, err)
				}
			}

		default:
			vf := reflect.ValueOf(i)
			if vf.Kind() != reflect.Map {
				return fmt.Errorf("expect the value is a map, but got %T", i)
			}

			for _, key := range vf.MapKeys() {
				if err := validator.Validate(c, key.Interface()); err != nil {
					return fmt.Errorf("map key '%v' is invalid: %v", key.Interface(), err)
				}
			}
		}

		return nil
	})
}

// MapV returns a new Validator to use the given validators to check
// each value of the map.
func MapV(validators ...Validator) Validator {
	if len(validators) == 0 {
		panic("MapVValidator: need at least one validator")
	}

	validator, desc := composeValidators("mapv", validators...)
	return NewValidator(desc, func(c, i interface{}) error {
		switch vs := i.(type) {
		case map[string]string:
			for _, value := range vs {
				if err := validator.Validate(c, value); err != nil {
					return fmt.Errorf("map value '%s' is invalid: %v", value, err)
				}
			}

		case map[string]interface{}:
			for _, value := range vs {
				if err := validator.Validate(c, value); err != nil {
					return fmt.Errorf("map value '%v' is invalid: %v", value, err)
				}
			}

		default:
			vf := reflect.ValueOf(i)
			if vf.Kind() != reflect.Map {
				return fmt.Errorf("expect the value is a map, but got %T", i)
			}

			for iter := vf.MapRange(); iter.Next(); {
				value := iter.Value().Interface()
				if err := validator.Validate(c, value); err != nil {
					return fmt.Errorf("map value '%v' is invalid: %v", value, err)
				}
			}
		}

		return nil
	})
}

// KV represents a key-value pair.
type KV struct {
	Key   interface{}
	Value interface{}
}

// MapKV returns a new Validator to use the given validators to check
// each key-value pair of the map.
//
// The value validated by the validators is a KV.
func MapKV(validators ...Validator) Validator {
	if len(validators) == 0 {
		panic("MapKVValidator: need at least one validator")
	}

	validator, desc := composeValidators("mapkv", validators...)
	return NewValidator(desc, func(c, i interface{}) error {
		switch vs := i.(type) {
		case map[string]string:
			for key, value := range vs {
				if err := validator.Validate(c, KV{Key: key, Value: value}); err != nil {
					return fmt.Errorf("map from key '%v' is invalid: %v", key, err)
				}
			}

		case map[string]interface{}:
			for key, value := range vs {
				if err := validator.Validate(c, KV{Key: key, Value: value}); err != nil {
					return fmt.Errorf("map from key '%v' is invalid: %v", key, err)
				}
			}

		default:
			vf := reflect.ValueOf(i)
			if vf.Kind() != reflect.Map {
				return fmt.Errorf("expect the value is a map, but got %T", i)
			}

			for iter := vf.MapRange(); iter.Next(); {
				key := iter.Key().Interface()
				value := iter.Value().Interface()
				if err := validator.Validate(c, KV{Key: key, Value: value}); err != nil {
					return fmt.Errorf("map from key '%v' is invalid: %v", key, err)
				}
			}
		}

		return nil
	})
}
