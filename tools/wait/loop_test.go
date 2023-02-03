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

package wait

import (
	"context"
	"testing"
	"time"
)

func TestUntil(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	cancel()

	Until(ctx, 0, ForeverRunner(func(context.Context) {
		t.Fatal("should not have been invoked")
	}))

	ctx, cancel = context.WithCancel(context.TODO())
	called := make(chan struct{})
	go func() {
		Until(ctx, 0, ForeverRunner(func(context.Context) { called <- struct{}{} }))
		close(called)
	}()
	<-called
	cancel()
	<-called
}

func TestNonSlidingUntil(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	cancel()
	SlidingUntil(ctx, 0, false, func(context.Context) (bool, error) {
		t.Fatal("should not have been invoked")
		return false, nil
	})

	ctx, cancel = context.WithCancel(context.TODO())
	called := make(chan struct{})
	go func() {
		SlidingUntil(ctx, 0, false, func(context.Context) (bool, error) {
			called <- struct{}{}
			return false, nil
		})
		close(called)
	}()
	<-called
	cancel()
	<-called
}

func TestUntilReturnsImmediately(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	now := time.Now()
	Until(ctx, 3*time.Second, ForeverRunner(func(context.Context) { cancel() }))
	if now.Add(2 * time.Second).Before(time.Now()) {
		t.Errorf("Until did not return immediately when the stop chan was closed inside the func")
	}
}

func TestJitterUntil(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	cancel()
	JitterUntil(ctx, 0, 1.0, func(context.Context) (bool, error) {
		t.Fatal("should not have been invoked")
		return false, nil
	})

	ctx, cancel = context.WithCancel(context.TODO())
	called := make(chan struct{})
	go func() {
		JitterUntil(ctx, 0, 1.0, func(context.Context) (bool, error) {
			called <- struct{}{}
			return false, nil
		})
		close(called)
	}()
	<-called
	cancel()
	<-called
}

func TestJitterUntilNegativeFactor(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	now := time.Now()
	called := make(chan struct{})
	received := make(chan struct{})
	go func() {
		JitterUntil(ctx, time.Second, -30.0, func(context.Context) (bool, error) {
			called <- struct{}{}
			<-received
			return false, nil
		})
	}()

	// first loop
	<-called
	received <- struct{}{}

	// second loop
	<-called
	cancel()
	received <- struct{}{}

	// it should take at most 2 seconds + some overhead, not 3
	if now.Add(3 * time.Second).Before(time.Now()) {
		t.Errorf("JitterUntil did not returned after predefined period with negative jitter factor when the stop chan was closed inside the func")
	}
}

func TestJitterLoopFirst(t *testing.T) {
	loop := NewJitterLoop(time.Second, time.Second, true, 0.0)

	start := time.Now()
	loop.Run(context.TODO(), func(ctx context.Context) (end bool, err error) {
		return true, nil
	})
	if duration := time.Since(start); duration < time.Second {
		t.Error("the first delay duration is too short")
	} else if duration > time.Second*2 {
		t.Error("the first delay duration is too long")
	}

	loop.StartDelay = -1
	start = time.Now()
	loop.Run(context.TODO(), func(ctx context.Context) (end bool, err error) {
		return true, nil
	})
	if duration := time.Since(start); duration < time.Second {
		t.Error("the run duration is too short")
	} else if duration > time.Second*2 {
		t.Error("the first delay duration is too long")
	}
}
