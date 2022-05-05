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

package task

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestService(t *testing.T) {
	runTask := func(f func(context.Context)) { Run(AsyncRunner(RunnerFunc(f))) }

	parent, pcancel := context.WithCancel(context.Background())
	DefaultService = NewService(parent)
	defer DefaultService.Deactivate()

	if Context() != nil {
		t.Errorf("unexpect the context")
	}
	if IsActivated() {
		t.Errorf("unexpect the task service is activated")
	}
	runTask(func(ctx context.Context) {
		t.Errorf("unexpect the task is run")
	})

	DefaultService.Activate()

	if Context() == nil {
		t.Errorf("expect the context, but got nil")
	}
	if !IsActivated() {
		t.Errorf("the task service is not activated")
	}

	var run atomic.Value
	runTask(func(ctx context.Context) { run.Store(true) })
	time.Sleep(time.Millisecond * 10)
	if v := run.Load(); v == nil || !v.(bool) {
		t.Errorf("the task is not run")
	}

	var err atomic.Value
	runTask(func(ctx context.Context) {
		<-ctx.Done()
		err.Store(ctx.Err())
	})

	DefaultService.Deactivate()
	time.Sleep(time.Millisecond * 10)
	if v := err.Load(); v == nil || v != context.Canceled {
		t.Errorf("expect the error context.Canceled, but got '%v'", v)
	}

	err = atomic.Value{}
	DefaultService.Activate()
	runTask(func(ctx context.Context) {
		<-ctx.Done()
		err.Store(ctx.Err())
	})

	pcancel()
	time.Sleep(time.Millisecond * 10)
	if v := err.Load(); v == nil || v != context.Canceled {
		t.Errorf("expect the error context.Canceled, but got '%s'", v)
	}

	if err := Context().Err(); err != context.Canceled {
		t.Errorf("expect the error context.Canceled, but got '%s'", err)
	}
}
