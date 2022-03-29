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

package leaderelection

import (
	"context"
	"time"
)

// ElectionRecord is the record information of the leader election.
type ElectionRecord struct {
	// HolderIdentity is the identity that owns the lease.
	//
	// If empty, no one owns this lease and all callers may acquire.
	// This value is set to empty when a client voluntarily steps down.
	HolderIdentity string `json:"holderIdentity" yaml:"holderIdentity"`

	LeaseDurationSeconds int `json:"leaseDurationSeconds" yaml:"leaseDurationSeconds"`
	LeaderTransitions    int `json:"leaderTransitions" yaml:"leaderTransitions"`

	AcquireTime time.Time `json:"acquireTime" yaml:"acquireTime"`
	RenewTime   time.Time `json:"renewTime" yaml:"renewTime"`
}

// Equal reports whether r is equal to other.
func (r ElectionRecord) Equal(other ElectionRecord) bool {
	return r.HolderIdentity == other.HolderIdentity &&
		r.LeaseDurationSeconds == other.LeaseDurationSeconds &&
		r.LeaderTransitions == other.LeaderTransitions &&
		r.AcquireTime.Equal(other.AcquireTime) &&
		r.RenewTime.Equal(other.RenewTime)
}

// ResourceLock offers a common interface for locking on arbitrary resources
// used in leader election.
type ResourceLock interface {
	Resource() string // String returns the resource that provides the locker.
	Get(context.Context) (record ElectionRecord, ok bool, err error)
	Create(context.Context, ElectionRecord) error
	Update(context.Context, ElectionRecord) error
}
