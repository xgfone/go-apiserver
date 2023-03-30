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

package handler_test

import (
	"fmt"
	"strconv"

	"github.com/xgfone/go-apiserver/tools/structfield"
	"github.com/xgfone/go-apiserver/tools/structfield/handler"
)

type _Int int

func (i *_Int) Set(v interface{}) error {
	*i = _Int(v.(int))
	return nil
}

type _Str string

func (s *_Str) Set(v interface{}) error {
	*s = _Str(v.(string))
	return nil
}

func ExampleNewSetterHandler() {
	// "set" is registered by default. Now, we register the customized
	// "setint" to pre-parse the tag value to int.
	structfield.Register("setint", handler.NewSetterHandler(func(s string) (interface{}, error) {
		i, err := strconv.ParseInt(s, 10, 64)
		return int(i), err
	}, nil))

	var t struct {
		Str _Str `set:"abc"`
		Int _Int `setint:"123"`
	}

	if err := structfield.Reflect(nil, &t); err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Int: %v\n", t.Int)
		fmt.Printf("Str: %v\n", t.Str)
	}

	// Output:
	// Int: 123
	// Str: abc
}
