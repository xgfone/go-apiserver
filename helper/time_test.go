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

package helper

import (
	"fmt"
	"time"
)

func ExampleTimeAdd() {
	t1 := time.Date(1, 2, 3, 4, 5, 6, 0, time.UTC)
	t2 := TimeAdd(t1, 6, 5, 4, 3, 2, 1)

	fmt.Printf("Year: %d\n", t2.Year())
	fmt.Printf("Month: %d\n", t2.Month())
	fmt.Printf("Day: %d\n", t2.Day())
	fmt.Printf("Hour: %d\n", t2.Hour())
	fmt.Printf("Minute: %d\n", t2.Minute())
	fmt.Printf("Second: %d\n", t2.Second())

	// Output:
	// Year: 7
	// Month: 7
	// Day: 7
	// Hour: 7
	// Minute: 7
	// Second: 7
}

func ExampleMustParseTime() {
	MustParseTime("2023-02-25 13:43:47", nil)                         // time.DateTime
	MustParseTime("2023-02-25T13:43:47+08:00", nil)                   // time.RFC3339
	MustParseTime("2023-02-25T13:43:47.123456+08:00", nil)            // time.RFC3339Nano
	MustParseTime("Sat, 25 Feb 2023 13:43:47 CST", nil, time.RFC1123) // time.RFC1123

	// Output:
}
