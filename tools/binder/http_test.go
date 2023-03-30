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

package binder

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
)

func ExampleBindStructFromStringMap() {
	src := map[string]string{
		"Int": "123",
		"Str": "456",
	}

	var dst struct {
		Int  int `tag:"-"`
		Int1 int `tag:"Int"`
		Int2 int `tag:"Str"`
	}

	err := BindStructFromStringMap(&dst, "tag", src)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Int=%d\n", dst.Int)
		fmt.Printf("Int1=%d\n", dst.Int1)
		fmt.Printf("Int2=%d\n", dst.Int2)
	}

	// Output:
	// Int=0
	// Int1=123
	// Int2=456
}

func ExampleBindStructFromHTTPHeader() {
	src := http.Header{
		"X-Int":  []string{"1", "2"},
		"X-Ints": []string{"3", "4"},
		"X-Str":  []string{"a", "b"},
		"X-Strs": []string{"c", "d"},
	}

	var dst struct {
		unexported string   `header:"-"`
		Other      string   `header:"Other"`
		Int        int      `header:"x-int"`
		Ints       []int    `header:"x-ints"`
		Str        string   `header:"x-str"`
		Strs       []string `header:"x-strs"`
	}

	err := BindStructFromHTTPHeader(&dst, "header", src)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("unexported=%s\n", dst.unexported)
		fmt.Printf("Other=%s\n", dst.Other)
		fmt.Printf("Int=%d\n", dst.Int)
		fmt.Printf("Ints=%d\n", dst.Ints)
		fmt.Printf("Str=%s\n", dst.Str)
		fmt.Printf("Strs=%s\n", dst.Strs)
	}

	// Output:
	// unexported=
	// Other=
	// Int=1
	// Ints=[3 4]
	// Str=a
	// Strs=[c d]
}

func ExampleBindStructFromURLValues() {
	src := url.Values{
		"int":  []string{"1", "2"},
		"ints": []string{"3", "4"},
		"str":  []string{"a", "b"},
		"strs": []string{"c", "d"},
	}

	var dst struct {
		unexported string   `qeury:"-"`
		Other      string   `query:"Other"`
		Int        int      `query:"int"`
		Ints       []int    `query:"ints"`
		Str        string   `query:"str"`
		Strs       []string `query:"strs"`
	}

	err := BindStructFromURLValues(&dst, "query", src)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("unexported=%s\n", dst.unexported)
		fmt.Printf("Other=%s\n", dst.Other)
		fmt.Printf("Int=%d\n", dst.Int)
		fmt.Printf("Ints=%d\n", dst.Ints)
		fmt.Printf("Str=%s\n", dst.Str)
		fmt.Printf("Strs=%s\n", dst.Strs)
	}

	// Output:
	// unexported=
	// Other=
	// Int=1
	// Ints=[3 4]
	// Str=a
	// Strs=[c d]
}

func ExampleBindStructFromMultipartFileHeaders() {
	src := map[string][]*multipart.FileHeader{
		"file":  {{Filename: "file"}},
		"files": {{Filename: "file1"}, {Filename: "file2"}},
		"_file": {{Filename: "file3"}},
	}

	var dst struct {
		Other       string                  `form:"Other"`
		_File       *multipart.FileHeader   `form:"_file"` // unexported, so ignored
		FileHeader  *multipart.FileHeader   `form:"file"`
		FileHeaders []*multipart.FileHeader `form:"files"`
	}

	err := BindStructFromMultipartFileHeaders(&dst, "form", src)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(dst.FileHeader.Filename)
		if dst._File != nil {
			fmt.Println(dst._File.Filename)
		}
		for _, fh := range dst.FileHeaders {
			fmt.Println(fh.Filename)
		}
	}

	// Output:
	// file
	// file1
	// file2
}
