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

//go:build go1.17
// +build go1.17

package helper

import (
	"context"
	"sync/atomic"
)

type cancelContext struct {
	context.Context
	context.CancelFunc
	reason atomic.Value
}

func (c *cancelContext) Err() error {
	if v := c.reason.Load(); v != nil {
		return v.(error)
	}
	return nil
}

func (c *cancelContext) cancel(reason error) {
	if reason == nil {
		panic("the context cancel reason must not be nil")
	}

	if c.reason.CompareAndSwap(nil, reason) {
		c.CancelFunc()
	}
}
