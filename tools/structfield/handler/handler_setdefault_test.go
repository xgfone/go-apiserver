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

package handler_test

import (
	"fmt"
	"time"

	"github.com/xgfone/go-apiserver/helper"
	"github.com/xgfone/go-apiserver/tools/structfield"
)

type defaultSetter string

func (d *defaultSetter) SetDefault(src interface{}) error {
	*d = defaultSetter(src.(string))
	return nil
}

func ExampleNewSetDefaultHandler() {
	// For test
	oldNow := helper.Now
	helper.Now = func() time.Time { return time.Unix(1660140928, 0).UTC() }
	defer func() { helper.Now = oldNow }()

	type String string
	type Struct struct {
		InnerInt int `default:"123"`
	}

	type S struct {
		Bool    bool    `default:"true"`
		Int     int     `default:"100"`
		Int8    int8    `default:"101"`
		Int16   int16   `default:"102"`
		Int32   int32   `default:"103"`
		Int64   int64   `default:"104"`
		Uint    uint    `default:"105"`
		Uint8   uint8   `default:"106"`
		Uint16  uint16  `default:"107"`
		Uint32  uint32  `default:"108"`
		Uint64  uint64  `default:"109"`
		Float32 float32 `default:"1.2"`
		Float64 float64 `default:"2.2"`
		String1 string  `default:"abc"`
		String2 String  `default:"xyz"`
		String3 string  `default:"xxx"`
		Struct  Struct
		Structs []Struct

		StructField int `default:".Struct.InnerInt"`

		TimeNowStr string `default:"now(2006-01-02T15:04:05Z)"`
		TimeNowInt int64  `default:"now()"`

		Setter      defaultSetter `default:"xyz"` // The type implementing helper.DefaultSetter
		DurationInt time.Duration `default:"1000"`
		DurationStr time.Duration `default:"2s"`
		TimeInt     time.Time     `default:"1658703387"` // 2022-07-24T22:56:27Z
		TimeStr     time.Time     `default:"2022-07-24T22:56:28Z"`

		IntPtr      *int           `default:"456"`
		TimePtr     *time.Time     `default:"2022-07-24T22:56:29Z"`
		DurationPtr *time.Duration `default:"3s"`
	}

	i := 123
	s := S{String3: "aaa", Structs: make([]Struct, 2), IntPtr: &i}
	err := structfield.Reflect(nil, &s) // NewSetDefaultHandler is registered into DefaultReflector.
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(s.Bool)
	fmt.Println(s.Int)
	fmt.Println(s.Int8)
	fmt.Println(s.Int16)
	fmt.Println(s.Int32)
	fmt.Println(s.Int64)
	fmt.Println(s.Uint)
	fmt.Println(s.Uint8)
	fmt.Println(s.Uint16)
	fmt.Println(s.Uint32)
	fmt.Println(s.Uint64)
	fmt.Println(s.Float32)
	fmt.Println(s.Float64)
	fmt.Println(s.String1)
	fmt.Println(s.String2)
	fmt.Println(s.String3)
	fmt.Println(s.Struct.InnerInt)
	fmt.Println(s.Structs[0].InnerInt)
	fmt.Println(s.Structs[1].InnerInt)
	fmt.Println(s.StructField)
	fmt.Println(s.TimeNowStr)
	fmt.Println(s.TimeNowInt)
	fmt.Println(s.Setter)
	fmt.Println(s.DurationInt)
	fmt.Println(s.DurationStr)
	fmt.Println(s.TimeInt.UTC().Format(time.RFC3339))
	fmt.Println(s.TimeStr.UTC().Format(time.RFC3339))
	fmt.Println(*s.IntPtr)
	fmt.Println(s.TimePtr.UTC().Format(time.RFC3339))
	fmt.Println(*s.DurationPtr)

	// Output:
	// true
	// 100
	// 101
	// 102
	// 103
	// 104
	// 105
	// 106
	// 107
	// 108
	// 109
	// 1.2
	// 2.2
	// abc
	// xyz
	// aaa
	// 123
	// 123
	// 123
	// 123
	// 2022-08-10T14:15:28Z
	// 1660140928
	// xyz
	// 1s
	// 2s
	// 2022-07-24T22:56:27Z
	// 2022-07-24T22:56:28Z
	// 123
	// 2022-07-24T22:56:29Z
	// 3s
}
