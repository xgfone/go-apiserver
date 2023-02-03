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

package result

import "testing"

func TestIsCode(t *testing.T) {
	if !IsCode("InstanceNotFound", "") {
		t.Errorf("expect the true, but got false")
	}

	if !IsCode("InstanceNotFound", "InstanceNotFound") {
		t.Errorf("expect true, but got false")
	}

	if IsCode("InstanceNotFound", "InstanceUnavailable") {
		t.Errorf("expect false, but got true")
	}

	if !IsCode("AuthFailure.TokenFailure", "AuthFailure") {
		t.Errorf("expect true, but got false")
	}

	if IsCode("AuthFailure", "AuthFailure.TokenFailure") {
		t.Errorf("expect false, but got true")
	}
}