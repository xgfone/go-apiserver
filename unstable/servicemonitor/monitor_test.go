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

package servicemonitor

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

type testService struct{ active int32 }

func newTestService() *testService       { return &testService{} }
func (s *testService) Activate()         { atomic.StoreInt32(&s.active, 1) }
func (s *testService) Deactivate()       { atomic.StoreInt32(&s.active, 0) }
func (s *testService) IsActivated() bool { return atomic.LoadInt32(&s.active) == 1 }

func TestMonitorCheckerExist(t *testing.T) {
	service := newTestService()
	checker := CheckerFunc(func(context.Context) (bool, error) { return true, nil })
	monitor := NewMonitor(service, checker, nil)

	monitor.Activate()
	time.Sleep(time.Millisecond * 100)
	if !monitor.IsActivated() {
		t.Error("the service monitor is not activated")
	}
	if !monitor.IsActive() {
		t.Error("the service is not activated")
	}
	if !service.IsActivated() {
		t.Error("the service is not activated")
	}

	monitor.Deactivate()
	time.Sleep(time.Millisecond * 100)
	if monitor.IsActivated() {
		t.Error("the service monitor is still activated")
	}
	if monitor.IsActive() {
		t.Error("the service is still activated")
	}
	if service.IsActivated() {
		t.Error("the service is still activated")
	}
}

func TestMonitorCheckerNotExist(t *testing.T) {
	service := newTestService()
	checker := CheckerFunc(func(context.Context) (bool, error) { return false, nil })
	monitor := NewMonitor(service, checker, nil)

	monitor.Activate()
	time.Sleep(time.Millisecond * 100)
	if !monitor.IsActivated() {
		t.Error("the service monitor is not activated")
	}
	if monitor.IsActive() {
		t.Error("the service is activated")
	}
	if service.IsActivated() {
		t.Error("the service is activated")
	}

	monitor.Deactivate()
	time.Sleep(time.Millisecond * 100)
	if monitor.IsActivated() {
		t.Error("the service monitor is still activated")
	}
}
