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

// Package signal provides a set of the signal handling functions.
package signal

import (
	"context"
	"os"
	"os/signal"
)

// ExitSignals is the set of the signals to let the program exit.
//
// For Unix/Linux or Windows, it contains the signals as follow:
//   syscall.SIGTERM
//   syscall.SIGQUIT
//   syscall.SIGABRT
//   syscall.SIGINT
var ExitSignals = []os.Signal{os.Interrupt}

// WaitExit monitors the exit signals and call the callback function
// to let the program exit.
func WaitExit(callback func()) { Once(Callback(callback), ExitSignals...) }

// Callback converts the function without arguments to the callback function.
func Callback(f func()) func(os.Signal) { return func(os.Signal) { f() } }

// Once monitors the given signals once and calls the callback function
// when any signal occurs.
func Once(callback func(os.Signal), signals ...os.Signal) {
	ch := make(chan os.Signal, 1)
	defer signal.Stop(ch)
	signal.Notify(ch, signals...)
	callback(<-ch)
}

// Loop loops to monitor the given signals and calls the callback function
// when any signal occurs unitl the context is done.
func Loop(c context.Context, cb func(os.Signal), sigs ...os.Signal) {
	ch := make(chan os.Signal, 1)
	defer signal.Stop(ch)
	signal.Notify(ch, sigs...)
	for {
		select {
		case <-c.Done():
			return
		case sig := <-ch:
			cb(sig)
		}
	}
}
