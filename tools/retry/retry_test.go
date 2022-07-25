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

package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func errcall(c context.Context) (bool, error) {
	return false, errors.New("test")
}

func TestNewPeriodicLoopRetry(t *testing.T) {
	retry := NewPeriodicLoopRetry(3, time.Millisecond*20)

	start := time.Now()
	if err := retry.Run(context.Background(), errcall); err == nil {
		t.Fail()
	} else if err.Error() != "test" {
		t.Errorf("the error is 'test': %s", err)
	}

	if cost := time.Since(start); cost < time.Millisecond*60 ||
		cost > time.Millisecond*120 {
		t.Error(cost)
	}

	start = time.Now()
	retry = NewPeriodicLoopRetry(5, 0)
	if err := retry.Run(context.TODO(), errcall); err == nil {
		t.Fail()
	} else if err.Error() != "test" {
		t.Errorf("the error is 'test': %s", err)
	}

	if cost := time.Since(start); cost > time.Millisecond*10 {
		t.Errorf("the cost of the retry call is greater than 10ms: %s", cost)
	}

	retry = NewPeriodicLoopRetry(2, time.Millisecond*100)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	retry.Run(ctx, func(context.Context) (bool, error) {
		t.Fatal("should not be called")
		return true, nil
	})

	err := retry.Run(context.TODO(), func(ctx context.Context) (bool, error) {
		return true, nil
	})
	if err != nil {
		t.Errorf("unexpected error '%s'", err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond*500)
	err = retry.Run(ctx, func(_ context.Context) (bool, error) {
		time.Sleep(time.Second)
		return false, errors.New("error")
	})
	if err != context.DeadlineExceeded {
		t.Errorf("expect error '%v', but got '%v'", context.DeadlineExceeded, err)
	}
	cancel()
}
