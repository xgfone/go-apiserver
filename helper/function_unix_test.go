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

//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package helper

import (
	"fmt"
	"os"
)

func ExampleFindCmd() {
	// Create a command file for test.
	os.Create("/tmp/test_cmd")
	defer os.Remove("/tmp/test_cmd")

	dirs := []string{
		"/bin",
		"/sbin",
		"/usr/bin",
		"/usr/sbin",
		"/usr/local/bin",
		"/usr/local/sbin",
		"/tmp",
	}

	cmd := FindCmd("test_cmd", dirs...)
	fmt.Println(cmd)

	// Output:
	// /tmp/test_cmd
}
