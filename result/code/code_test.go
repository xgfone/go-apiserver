// Copyright 2024 xgfone
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

package code

import "testing"

func TestIs(t *testing.T) {
	if !ErrIs(ErrBadRequest, "BadRequest") {
		t.Error("fail")
	}

	if !ErrBadRequest.IsBadRequest() {
		t.Error("fail")
	}
	if ErrBadRequest.IsAuthFailure() || ErrBadRequest.IsUnallowed() ||
		ErrBadRequest.IsNotFound() || ErrBadRequest.IsUnauthorized() ||
		ErrBadRequest.IsInternalServerError() {
		t.Error("fail")
	}

	if !Is("InstanceNotFound", "") {
		t.Errorf("expect the true, but got false")
	}

	if !Is("InstanceNotFound", "InstanceNotFound") {
		t.Errorf("expect true, but got false")
	}

	if Is("InstanceNotFound", "InstanceUnavailable") {
		t.Errorf("expect false, but got true")
	}

	if !Is("AuthFailure.TokenFailure", "AuthFailure") {
		t.Errorf("expect true, but got false")
	}

	if Is("AuthFailure", "AuthFailure.TokenFailure") {
		t.Errorf("expect false, but got true")
	}

	if !Is(123, 123) {
		t.Errorf("expect true, but got false")
	}

	if Is(123, 456) {
		t.Errorf("expect false, but got true")
	}

	if Is(int64(123), uint64(123)) { // Because the types of them are not the same.
		t.Errorf("expect false, but got true")
	}
}
