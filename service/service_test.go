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
	"testing"
)

type testService struct{ name string }

func newTestService(name string) testService { return testService{name} }
func (s testService) Activate()              {}
func (s testService) Deactivate()            {}

func TestServices(t *testing.T) {
	origs := Services{newTestService("svc1"), newTestService("svc2")}

	svcs1 := origs.Append(newTestService("svc3"))
	svcs2 := origs.Clone(newTestService("svc4"))

	for _, svc := range svcs1 {
		switch name := svc.(testService).name; name {
		case "svc1", "svc2", "svc3":
		default:
			t.Errorf("unexpected service named '%s'", name)
		}
	}

	for _, svc := range svcs2 {
		switch name := svc.(testService).name; name {
		case "svc1", "svc2", "svc4":
		default:
			t.Errorf("unexpected service named '%s'", name)
		}
	}

	origs.Activate()
	origs.Deactivate()
}
