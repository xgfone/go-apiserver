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

// Package wait provides some tools to loop calling until a condition reaches.
package wait

import (
	"context"
	"sync"

	"github.com/xgfone/go-apiserver/log"
)

func handlePanic(r interface{}) {
	log.Error("wrap a panic", "panic", r)
}

func wrapPanic() {
	if r := recover(); r != nil {
		handlePanic(r)
	}
}

func safeRun(c context.Context, r Runner) (bool, error) {
	defer wrapPanic()
	return r(c)
}

// HandlePanic is used to handle the panic.
var HandlePanic func(interface{}) = handlePanic

// Group allows to start a group of goroutines and wait for their completion.
type Group struct{ wg sync.WaitGroup }

// Wait waits until all the goroutines end.
func (g *Group) Wait() { g.wg.Wait() }

// StartWithChannel starts f in a new goroutine in the group.
// stopCh is passed to f as an argument. f should stop when stopCh is available.
func (g *Group) StartWithChannel(stopCh <-chan struct{}, f func(stopCh <-chan struct{})) {
	g.Start(func() { f(stopCh) })
}

// StartWithContext starts f in a new goroutine in the group.
// ctx is passed to f as an argument. f should stop when ctx.Done() is available.
func (g *Group) StartWithContext(ctx context.Context, f func(context.Context)) {
	g.Start(func() { f(ctx) })
}

// Start starts f in a new goroutine in the group.
func (g *Group) Start(f func()) {
	g.wg.Add(1)
	go func() { defer g.wg.Done(); f() }()
}
