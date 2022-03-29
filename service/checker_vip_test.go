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

package service

import (
	"context"
	"testing"
)

func TestVipCheckerExist(t *testing.T) {
	checker := NewVipChecker("127.0.0.1", "")
	ok, err := checker.Check(context.Background())
	if err != nil {
		t.Error(err)
	} else if !ok {
		t.Error("expect ip '127.0.0.1' exists, but got none")
	}
}

func TestVipCheckerNotExist(t *testing.T) {
	checker := NewVipChecker("1.2.3.4", "")
	ok, err := checker.Check(context.Background())
	if err != nil {
		t.Error(err)
	} else if ok {
		t.Error("unexpect ip '127.0.0.1' exists, but got")
	}
}
