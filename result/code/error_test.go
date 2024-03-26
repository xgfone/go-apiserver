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

import (
	"errors"
	"testing"
)

func TestError(t *testing.T) {
	e := NewError("Ok", "").WithError(errors.New("test")).
		WithCtx(200).WithMessage("").WithMessage("%s", "msg")

	expect := "Ok: msg"
	if s := e.Error(); s != expect {
		t.Errorf("expect '%s', but got '%s'", expect, s)
	}

	expect = "code=Ok, msg=msg"
	if s := e.String(); s != expect {
		t.Errorf("expect '%s', but got '%s'", expect, s)
	}

	if err := errors.Unwrap(e); err == nil {
		t.Error("expect an error, but got nil")
	} else if s := err.Error(); s != "test" {
		t.Errorf("expect error '%s', but got '%s'", "test", s)
	}

	_ = e.GetCode()
}
