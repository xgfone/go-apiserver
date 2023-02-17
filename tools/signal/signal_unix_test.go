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

//go:build unix

package signal

import (
	"context"
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestSignal(t *testing.T) {
	var i int32
	go Once(context.Background(), Callback(func() { atomic.StoreInt32(&i, 1) }), syscall.SIGHUP)

	time.Sleep(time.Millisecond * 50)
	Kill(os.Getpid(), syscall.SIGHUP)
	time.Sleep(time.Millisecond * 50)

	if v := atomic.LoadInt32(&i); v != 1 {
		t.Errorf("expect %d, bug got %d", 1, v)
	}
}
