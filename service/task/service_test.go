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
	"testing"
	"time"
)

func TestService(t *testing.T) {
	parent, pcancel := context.WithCancel(context.Background())
	DefaultService = NewService(parent)
	defer DefaultService.Deactivate()

	if Context() != nil {
		t.Errorf("unexpect the context")
	}
	if IsActivated() {
		t.Errorf("unexpect the task service is activated")
	}
	Run("id", "name", func(ctx context.Context) {
		t.Errorf("unexpect the task is run")
	})

	DefaultService.Activate()

	if Context() == nil {
		t.Errorf("expect the context, but got nil")
	}
	if !IsActivated() {
		t.Errorf("the task service is not activated")
	}

	var run bool
	Run("id", "name", func(ctx context.Context) { run = true })
	time.Sleep(time.Millisecond * 10)
	if !run {
		t.Errorf("the task is not run")
	}

	var err error
	Run("id", "name", func(ctx context.Context) {
		<-ctx.Done()
		err = ctx.Err()
	})

	DefaultService.Deactivate()
	time.Sleep(time.Millisecond * 10)
	if err != context.Canceled {
		t.Errorf("expect the error context.Canceled, but got '%s'", err.Error())
	}

	err = nil
	DefaultService.Activate()
	Run("id", "name", func(ctx context.Context) {
		<-ctx.Done()
		err = ctx.Err()
	})

	pcancel()
	time.Sleep(time.Millisecond * 10)
	if err != context.Canceled {
		t.Errorf("expect the error context.Canceled, but got '%s'", err.Error())
	}

	if err = Context().Err(); err != context.Canceled {
		t.Errorf("expect the error context.Canceled, but got '%s'", err.Error())
	}
}
