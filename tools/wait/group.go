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
	"sync"

	"github.com/xgfone/go-apiserver/log"
)

// Group allows to start a group of goroutines and wait for their completion.
type Group struct{ wg sync.WaitGroup }

// Wait waits until all the goroutines end.
func (g *Group) Wait() { g.wg.Wait() }

// GoWithChannel starts to run the function f in a new goroutine in the group.
// stopCh is passed to f as an argument. f should stop when stopCh is available.
func (g *Group) GoWithChannel(stopCh <-chan struct{}, f func(stopCh <-chan struct{})) {
	g.Go(func() { f(stopCh) })
}

// GoWithContext starts to run the function f in a new goroutine in the group.
// ctx is passed to f as an argument. f should stop when ctx.Done() is available.
func (g *Group) GoWithContext(ctx context.Context, f func(context.Context)) {
	g.Go(func() { f(ctx) })
}

// Go starts to run the function f in a new goroutine in the group.
func (g *Group) Go(f func()) {
	g.wg.Add(1)
	go g.run(f)
}

func (g *Group) run(f func()) {
	defer g.wg.Done()
	defer log.WrapPanic()
	f()
}
