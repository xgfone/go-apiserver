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

package test

import (
	"slices"
	"testing"
)

// CheckStatusCode checks whether the result status code is equal to the expect.
func CheckStatusCode(t *testing.T, name string, result, expect int) {
	if result != expect {
		t.Errorf("%s: expect the status code '%d', but got '%d'",
			name, expect, result)
	}
}

// CheckStrings checks whether the two strings are exactly equal.
func CheckStrings(t *testing.T, name string, results, expects []string) {
	if len(results) != len(expects) {
		t.Errorf("%s: expect %d lines, but got %d: %v",
			name, len(expects), len(results), results)
		return
	}

	for i := 0; i < len(results); i++ {
		if results[i] != expects[i] {
			t.Errorf("%s: expect '%s', but got '%s'", name, expects[i], results[i])
		}
	}
}

// InStrings checks whether each of the result strings is in the expect strings.
func InStrings(t *testing.T, name string, results, expects []string) {
	for _, s := range results {
		if !slices.Contains(expects, s) {
			t.Errorf("%s: the result string '%s' is not in %v", name, s, expects)
		}
	}
}
