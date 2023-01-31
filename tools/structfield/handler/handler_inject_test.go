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

var (
	_ handler.Injector = new(_Int)
	_ handler.Injector = new(_Str)
)

type _Int int

func (i *_Int) Inject(v interface{}) error {
	*i = _Int(v.(int))
	return nil
}

type _Str string

func (s *_Str) Inject(v interface{}) error {
	*s = _Str(v.(string))
	return nil
}

func ExampleNewInjectHandler() {
	// "inject" is registered by default. Now, we register the customized
	// "injectint" to pre-parse the tag value to int.
	structfield.Register("injectint", handler.NewInjectHandler(func(s string) (interface{}, error) {
		i, err := strconv.ParseInt(s, 10, 64)
		return int(i), err
	}))

	var t struct {
		Str _Str `inject:"abc"`
		Int _Int `injectint:"123"`
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
