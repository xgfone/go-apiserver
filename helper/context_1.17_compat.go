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

//go:build !go1.17
// +build !go1.17

package helper

import (
	"context"
	"sync"
)

type cancelContext struct {
	context.Context
	context.CancelFunc

	lock sync.RWMutex
	err  error
}

func (c *cancelContext) Err() (err error) {
	c.lock.RLock()
	err = c.err
	c.lock.RUnlock()
	return
}

func (c *cancelContext) cancel(reason error) {
	if reason == nil {
		panic("the context cancel reason must not be nil")
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	if c.err == nil {
		c.err = reason
		c.CancelFunc()
	}
}
