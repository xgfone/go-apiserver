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

package validators

import "testing"

func TestValidatorString(t *testing.T) {
	testString(t, Cidr().String(), `cidr`)
	testString(t, IP().String(), `ip`)
	testString(t, Mac().String(), `mac`)
	testString(t, Max(123).String(), `max(123)`)
	testString(t, Min(123).String(), `min(123)`)
	testString(t, OneOf("a", "b").String(), `oneof("a","b")`)
	testString(t, Required().String(), `required`)
	testString(t, Zero().String(), `zero`)
}

func testString(t *testing.T, result, expect string) {
	if result != expect {
		t.Errorf("expect validator '%s', but got '%s'", expect, result)
	}
}
