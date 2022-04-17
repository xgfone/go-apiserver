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

// Package validation provides a validation framework based on the rule.
package validation

import (
	"fmt"
	"reflect"
	"strings"
)

// Validator is a validator to check whether a value is valid.
type Validator interface {
	Validate(i interface{}) error
	String() string
}

// NewValidator returns the new Validator based on the validation rule
// and function.
func NewValidator(rule string, validate func(interface{}) error) Validator {
	return validator{s: rule, f: validate}
}

type validator struct {
	s string
	f func(interface{}) error
}

func (v validator) Validate(i interface{}) error { return v.f(i) }
func (v validator) String() string               { return v.s }

func formatValidators(name string, validators []Validator) string {
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
			b.WriteString(name)
		}
		b.WriteString(validator.String())
	}
	b.WriteByte(')')

	return b.String()
}

// ************************************************************************* //

// AndValidator is a And validator based on a set of the validators.
type andValidator []Validator

// Validate implements the interface Validator.
func (vs andValidator) Validate(v interface{}) (err error) {
	for i, _len := 0, len(vs); i < _len; i++ {
		if err = vs[i].Validate(v); err != nil {
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
func (vs orValidator) Validate(v interface{}) (err error) {
	for i, _len := 0, len(vs); i < _len; i++ {
		if err = vs[i].Validate(v); err == nil {
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

	validator := And(validators...)
	desc := fmt.Sprintf("array(%s)", validator.String())
	return NewValidator(desc, func(i interface{}) error {
		switch vs := i.(type) {
		case []string:
			for i, s := range vs {
				if err := validator.Validate(s); err != nil {
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
				if err := validator.Validate(vf.Index(i).Interface()); err != nil {
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

	validator := And(validators...)
	desc := fmt.Sprintf("mapk(%s)", validator.String())
	return NewValidator(desc, func(i interface{}) error {
		switch vs := i.(type) {
		case map[string]string:
			for key := range vs {
				if err := validator.Validate(key); err != nil {
					return fmt.Errorf("map key '%s' is invalid: %v", key, err)
				}
			}

		case map[string]interface{}:
			for key := range vs {
				if err := validator.Validate(key); err != nil {
					return fmt.Errorf("map key '%s' is invalid: %v", key, err)
				}
			}

		default:
			vf := reflect.ValueOf(i)
			if vf.Kind() != reflect.Map {
				return fmt.Errorf("expect the value is a map, but got %T", i)
			}

			for _, key := range vf.MapKeys() {
				if err := validator.Validate(key.Interface()); err != nil {
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

	validator := And(validators...)
	desc := fmt.Sprintf("mapv(%s)", validator.String())
	return NewValidator(desc, func(i interface{}) error {
		switch vs := i.(type) {
		case map[string]string:
			for _, value := range vs {
				if err := validator.Validate(value); err != nil {
					return fmt.Errorf("map value '%s' is invalid: %v", value, err)
				}
			}

		case map[string]interface{}:
			for _, value := range vs {
				if err := validator.Validate(value); err != nil {
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
				if err := validator.Validate(value); err != nil {
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

	validator := And(validators...)
	desc := fmt.Sprintf("mapkv(%s)", validator.String())
	return NewValidator(desc, func(i interface{}) error {
		switch vs := i.(type) {
		case map[string]string:
			for key, value := range vs {
				if err := validator.Validate(KV{Key: key, Value: value}); err != nil {
					return fmt.Errorf("map from key '%v' is invalid: %v", key, err)
				}
			}

		case map[string]interface{}:
			for key, value := range vs {
				if err := validator.Validate(KV{Key: key, Value: value}); err != nil {
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
				if err := validator.Validate(KV{Key: key, Value: value}); err != nil {
					return fmt.Errorf("map from key '%v' is invalid: %v", key, err)
				}
			}
		}

		return nil
	})
}
