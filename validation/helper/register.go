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

// Package helper provides some helpful functions to register the builder function.
package helper

import (
	"fmt"

	"github.com/xgfone/go-apiserver/validation"
	"github.com/xgfone/predicate"
)

// RegisterFuncNoArg is used to help to register the builder function
// into builder to parse the validator without any arguments.
func RegisterFuncNoArg(b *validation.Builder, name string, newf func() validation.Validator) {
	b.RegisterFunc(name, func(c *validation.Context, args ...interface{}) (err error) {
		if len(args) > 0 {
			err = fmt.Errorf("%s must not have any arguments", name)
		} else {
			c.AppendValidators(newf())
		}
		return
	})
}

// RegisterFuncOneFloat is used to help to register the builder function
// into builder to parse the validator with only one float64 argument.
func RegisterFuncOneFloat(b *validation.Builder, name string, newf func(float64) validation.Validator) {
	b.RegisterFunc(name, func(c *validation.Context, args ...interface{}) (err error) {
		if len(args) != 1 {
			return fmt.Errorf("%s must have and only have one argument", name)
		}

		v, err := getFloat(name, -1, args[0])
		if err == nil {
			c.AppendValidators(newf(v))
		}

		return
	})
}

func getFloat(name string, index int, i interface{}) (f float64, err error) {
	switch v := i.(type) {
	case int:
		f = float64(v)

	case float64:
		f = v

	default:
		if index < 0 {
			err = fmt.Errorf("%s does not support the argument type %T", name, i)
		} else {
			err = fmt.Errorf("%s expects %dth argument is an int or float, but got %T", name, index, i)
		}
	}

	return
}

// RegisterFuncTwoFloats is used to help to register the builder function
// into builder to parse the validator with only two float64 arguments.
func RegisterFuncTwoFloats(b *validation.Builder, name string, newf func(float64, float64) validation.Validator) {
	b.RegisterFunc(name, func(c *validation.Context, args ...interface{}) (err error) {
		if len(args) != 2 {
			return fmt.Errorf("%s must have and only have two argument", name)
		}

		first, err := getFloat(name, 0, args[0])
		if err != nil {
			return
		}

		second, err := getFloat(name, 1, args[1])
		if err != nil {
			return
		}

		c.AppendValidators(newf(first, second))
		return
	})
}

// RegisterFuncFloats is used to help to register the builder function
// into builder to parse the validator with a set of float64 arguments.
func RegisterFuncFloats(b *validation.Builder, name string, newf func(...float64) validation.Validator) {
	b.RegisterFunc(name, func(c *validation.Context, args ...interface{}) (err error) {
		vs := make([]float64, len(args))
		for i, v := range args {
			if vs[i], err = getFloat(name, i, v); err != nil {
				return err
			}
		}

		c.AppendValidators(newf(vs...))
		return
	})
}

// RegisterFuncStrings is used to help to register the builder function
// into builder to parse the validator with a set of strings.
func RegisterFuncStrings(b *validation.Builder, name string, newf func(...string) validation.Validator) {
	b.RegisterFunc(name, func(c *validation.Context, args ...interface{}) (err error) {
		var ok bool
		ss := make([]string, len(args))
		for i, v := range args {
			if ss[i], ok = v.(string); !ok {
				return fmt.Errorf("%s expects %dth argument is a string, but got %T", name, i, v)
			}
		}
		c.AppendValidators(newf(ss...))
		return
	})
}

// RegisterFuncValidators is used to help to register the builder function
// into builder to parse the validator with a set of other validators.
func RegisterFuncValidators(b *validation.Builder, name string,
	newf func(...validation.Validator) validation.Validator) {
	b.RegisterFunc(name, func(c *validation.Context, args ...interface{}) (err error) {
		if len(args) == 0 {
			return fmt.Errorf("%s validator has no argument", name)
		}

		ac := c.New()
		for i, arg := range args {
			b, ok := arg.(predicate.ContextBuilder)
			if !ok {
				return fmt.Errorf("%s expects %dth argument is a validator, but got %T", name, i, arg)
			}

			nc := ac.New()
			if err := b.Build(nc); err != nil {
				return err
			}
			ac.And(nc)
		}

		c.AppendValidators(newf(ac.(*validation.Context).Validators()...))
		return
	})
}
