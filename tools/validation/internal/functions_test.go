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

package internal

import "testing"

func TestOneOf(t *testing.T) {
	oneof := NewOneOf("oneof", "a", "b", "c")

	if err := oneof.Validate(nil, "0"); err == nil {
		t.Errorf("expect an error, but got nil")
	}

	if err := oneof.Validate(nil, "b"); err != nil {
		t.Errorf("expect nil, but got an error: %v", err)
	}
}
