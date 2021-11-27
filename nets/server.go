// Copyright 2021 xgfone
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

package nets

import "sync/atomic"

// RuntimeState is the runtime information of a service.
type RuntimeState struct {
	Total   uint64 // The total number to handle all the requests.
	Success uint64 // The total number to handle the requests successfully.
	Current uint64 // The number of the requests that are being handled.
}

// IncSuccess increases the success state.
func (rs *RuntimeState) IncSuccess() {
	atomic.AddUint64(&rs.Success, 1)
}

// Inc increases the total and current state.
func (rs *RuntimeState) Inc() {
	atomic.AddUint64(&rs.Total, 1)
	atomic.AddUint64(&rs.Current, 1)
}

// Dec decreases the current state.
func (rs *RuntimeState) Dec() {
	atomic.AddUint64(&rs.Current, ^uint64(0))
}

// Clone clones itself to a new one.
func (rs *RuntimeState) Clone() RuntimeState {
	return RuntimeState{
		Total:   atomic.LoadUint64(&rs.Total),
		Success: atomic.LoadUint64(&rs.Success),
		Current: atomic.LoadUint64(&rs.Current),
	}
}
