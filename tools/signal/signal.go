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

package signal

import (
	"os"
	"os/signal"
)

// DefaultSignals is the set of the default signals.
//
// For Unix/Linux or Windows, it contains the signals as follow:
//   syscall.SIGTERM
//   syscall.SIGQUIT
//   syscall.SIGABRT
//   syscall.SIGINT
var DefaultSignals = []os.Signal{os.Interrupt}

// Signal monitors the given signals and calls the callback function
// when any signal occurs.
//
// If no signals are given, use DefaultSignals instead.
func Signal(callback func(), signals ...os.Signal) {
	if len(signals) == 0 {
		if len(DefaultSignals) == 0 {
			panic("no signals to be monitored")
		}
		signals = DefaultSignals
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)

	<-ch
	callback()
}
