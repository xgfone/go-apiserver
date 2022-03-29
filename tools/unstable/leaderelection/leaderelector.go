/*
Copyright 2015 The Kubernetes Authors.
Copyright 2022 xgfone

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package leaderelection implements leader election of a set of nodes.
//
// A client only acts on timestamps captured locally to infer the state of the
// leader election. The client does not consider timestamps in the leader
// election record to be accurate because these timestamps may not have been
// produced by a local clock. The implemention does not depend on their
// accuracy and only uses their change to indicate that another client has
// renewed the leader lease. Thus the implementation is tolerant to arbitrary
// clock skew, but is not tolerant to arbitrary clock skew rate.
//
// However the level of tolerance to skew rate can be configured by setting
// RenewDeadline and LeaseDuration appropriately. The tolerance expressed as a
// maximum tolerated ratio of time passed on the fastest node to time passed on
// the slowest node can be approximately achieved with a configuration that sets
// the same ratio of LeaseDuration to RenewDeadline. For example if a user wanted
// to tolerate some nodes progressing forward in time twice as fast as other nodes,
// the user could set LeaseDuration to 60 seconds and RenewDeadline to 30 seconds.
//
// While not required, some method of clock synchronization between nodes in the
// cluster is highly recommended. It's important to keep in mind when configuring
// this client that the tolerance to skew rate varies inversely to master
// availability.
//
// Larger clusters often have a more lenient SLA for API latency. This should be
// taken into account when configuring the client. The rate of leader transitions
// should be monitored and RetryPeriod and LeaseDuration should be increased
// until the rate is stable and acceptably low. It's important to keep in mind
// when configuring this client that the tolerance to API latency varies inversely
// to master availability.
//
// Notice: it is ported from the package "k8s.io/client-go@v0.23.5/tools/leaderelection".
package leaderelection

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/tools/wait"
)

// LeaderCallbacks are callbacks that are triggered during certain lifecycle
// events of the LeaderElector. These are invoked asynchronously.
type LeaderCallbacks struct {
	// OnStartedLeading is called when a LeaderElector client starts leading.
	OnStartedLeading func(context.Context)

	// OnStoppedLeading is called when a LeaderElector client stops leading.
	OnStoppedLeading func()

	// OnNewLeader is called when the client observes a leader
	// that is not the previously observed leader.
	//
	// This includes the first observed leader when the client starts.
	OnNewLeader func(identity string)
}

// ElectionConfig is used to configure the leader election.
type ElectionConfig struct {
	// Identity indicates the identity of the elector.
	Identity string

	// Lock is the resource that will be used for locking.
	Lock ResourceLock

	// LeaseDuration is the duration that non-leader candidates will wait
	// to force acquire leadership. This is measured against time of last
	// observed ack.
	//
	// A client needs to wait a full LeaseDuration without observing a change
	// to the record before it can attempt to take over. When all clients are
	// shutdown and a new set of clients are started with different names
	// against the same leader record, they must wait the full LeaseDuration
	// before attempting to acquire the lease. Thus LeaseDuration should be
	// as short as possible (within your tolerance for clock skew rate) to
	// avoid a possible long waits in the scenario.
	//
	// Default: 10s
	LeaseDuration time.Duration

	// RenewDeadline is the duration that the acting master will retry
	// refreshing leadership before giving up.
	//
	// Default: 6s
	RenewDeadline time.Duration

	// RetryPeriod is the duration the LeaderElector clients should wait
	// between tries of actions.
	//
	// Default: 2s
	RetryPeriod time.Duration

	// Callbacks are callbacks that are triggered during certain lifecycle
	// events of the LeaderElector.
	Callbacks LeaderCallbacks

	// ReleaseOnCancel should be set true if the lock should be released
	// when the run context is cancelled.
	//
	// If you set this to true, you must ensure all code guarded by this lease
	// has successfully completed prior to cancelling the context, or you may
	// have two processes simultaneously acting on the critical path.
	ReleaseOnCancel bool

	// Name is the name of the resource lock for debugging.
	Name string
}

// LeaderElector is a leader election client.
type LeaderElector struct {
	config         ElectionConfig
	observedRecord atomic.Value
	reportedLeader string
}

// NewLeaderElector creates a LeaderElector from a LeaderElectionConfig
func NewLeaderElector(config ElectionConfig) (*LeaderElector, error) {
	if config.LeaseDuration == 0 {
		config.LeaseDuration = time.Second * 10
	}
	if config.RenewDeadline == 0 {
		config.RenewDeadline = time.Second * 6
	}
	if config.RetryPeriod == 0 {
		config.RetryPeriod = time.Second * 2
	}

	if config.LeaseDuration <= config.RenewDeadline {
		return nil, fmt.Errorf("leaseDuration must be greater than renewDeadline")
	}
	if config.RenewDeadline <= time.Duration(1.2*float64(config.RetryPeriod)) {
		return nil, fmt.Errorf("renewDeadline must be greater than retryPeriod*1.2")
	}
	if config.LeaseDuration < 1 {
		return nil, fmt.Errorf("leaseDuration must be greater than zero")
	}
	if config.RenewDeadline < 1 {
		return nil, fmt.Errorf("renewDeadline must be greater than zero")
	}
	if config.RetryPeriod < 1 {
		return nil, fmt.Errorf("retryPeriod must be greater than zero")
	}
	if config.Callbacks.OnStartedLeading == nil {
		return nil, fmt.Errorf("OnStartedLeading callback must not be nil")
	}
	if config.Callbacks.OnStoppedLeading == nil {
		return nil, fmt.Errorf("OnStoppedLeading callback must not be nil")
	}

	if config.Lock == nil {
		return nil, fmt.Errorf("Lock must not be nil")
	}
	if config.Identity == "" {
		return nil, fmt.Errorf("elector identity must not be empty")
	}

	le := &LeaderElector{config: config}
	le.setObservedRecord(ElectionRecord{})
	return le, nil
}

// Run starts the leader election loop, which will not return
// before leader election loop is stopped by ctx.
func (le *LeaderElector) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			le.run(ctx)
		}
	}
}

// RunOnce starts the leader election loop, which will not return
// before leader election loop is stopped by ctx or it has stopped
// holding the leader lease.
func (le *LeaderElector) RunOnce(ctx context.Context) {
	le.run(ctx)
}

func (le *LeaderElector) run(ctx context.Context) {
	defer log.WrapPanic()
	defer func() { le.config.Callbacks.OnStoppedLeading() }()

	if !le.acquire(ctx) {
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go le.config.Callbacks.OnStartedLeading(ctx)
	le.renew(ctx)
}

// GetLeader returns the identity of the last observed leader
// or returns the empty string if no leader has yet been observed.
func (le *LeaderElector) GetLeader() string {
	rcd, _ := le.getObservedRecord()
	return rcd.HolderIdentity
}

// IsLeader returns true if the last observed leader was this client
// else returns false.
func (le *LeaderElector) IsLeader() bool {
	rcd, _ := le.getObservedRecord()
	return le.isLeader(rcd)
}

func (le *LeaderElector) isLeader(rcd ElectionRecord) bool {
	return rcd.HolderIdentity == le.config.Identity
}

func (le *LeaderElector) info(msg string) {
	rsc := le.config.Lock.Resource()
	log.Log(log.LvlInfo, 1, msg, "elector", le.config.Name, "resource", rsc)
}

func (le *LeaderElector) error(msg string, err error) {
	rsc := le.config.Lock.Resource()
	log.Log(log.LvlError, 1, msg, "elector", le.config.Name, "resource", rsc, "err", err)
}

// acquire loops calling tryAcquireOrRenew and returns true immediately
// when tryAcquireOrRenew succeeds. Returns false if ctx signals done.
func (le *LeaderElector) acquire(ctx context.Context) (succeeded bool) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	le.info("attempting to acquire leader lease")
	wait.JitterUntil(func() {
		succeeded = le.tryAcquireOrRenew(ctx)
		le.maybeReportTransition()
		if succeeded {
			cancel()
			le.info("successfully acquire lease")
		}
	}, le.config.RetryPeriod, 1.2, true, ctx.Done())
	return
}

// renew loops calling tryAcquireOrRenew and returns immediately
// when tryAcquireOrRenew fails or ctx signals done.
func (le *LeaderElector) renew(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wait.Until(func() {
		timeoutCtx, timeoutCancel := context.WithTimeout(ctx, le.config.RenewDeadline)
		defer timeoutCancel()

		err := wait.PollImmediateUntil(le.config.RetryPeriod, func() (bool, error) {
			return le.tryAcquireOrRenew(timeoutCtx), nil
		}, timeoutCtx.Done())

		le.maybeReportTransition()
		if err == nil {
			le.info("successfully renewed lease")
		} else {
			le.error("fail to renew lease", err)
			cancel()
		}
	}, le.config.RetryPeriod, ctx.Done())

	// If we hold the lease, give it up
	if le.config.ReleaseOnCancel {
		le.release()
	}
}

// release attempts to release the leader lease if we have acquired it.
func (le *LeaderElector) release() bool {
	if !le.IsLeader() {
		return true
	}

	now := time.Now()
	observedRecord, _ := le.getObservedRecord()
	electionRecord := ElectionRecord{
		LeaderTransitions:    observedRecord.LeaderTransitions,
		LeaseDurationSeconds: 1,
		AcquireTime:          now,
		RenewTime:            now,
	}
	if err := le.config.Lock.Update(context.TODO(), electionRecord); err != nil {
		le.error("fail to release the resource lock", err)
		return false
	}

	le.setObservedRecord(electionRecord)
	return true
}

// tryAcquireOrRenew tries to acquire a leader lease if it is not already
// acquired, else it tries to renew the lease if it has already been acquired.
//
// Returns true on success else returns false.
func (le *LeaderElector) tryAcquireOrRenew(ctx context.Context) (success bool) {
	now := time.Now()
	electionRecord := ElectionRecord{
		HolderIdentity:       le.config.Identity,
		LeaseDurationSeconds: int(le.config.LeaseDuration / time.Second),
		RenewTime:            now,
		AcquireTime:          now,
	}

	// 1. obtain or create the ElectionRecord
	oldElectionRecord, ok, err := le.config.Lock.Get(ctx)
	if err != nil {
		le.error("fail to retrieve resource lock", err)
		return false
	} else if !ok { // NotFound
		if err = le.config.Lock.Create(ctx, electionRecord); err != nil {
			le.error("fail to initially create leader election record", err)
			return false
		}

		le.setObservedRecord(electionRecord)
		return true
	}

	// 2. Record obtained, check the Identity & Time
	observedRecord, observedTime := le.getObservedRecord()
	if !observedRecord.Equal(oldElectionRecord) {
		le.setObservedRecord(oldElectionRecord)
		observedRecord, observedTime = oldElectionRecord, now
	}
	if !le.isLeader(observedRecord) &&
		len(oldElectionRecord.HolderIdentity) > 0 &&
		observedTime.Add(le.config.LeaseDuration).After(now) {
		log.Debug("resource lock is held and has not yet expired",
			"holder", oldElectionRecord.HolderIdentity)
		return false
	}

	// 3. We're going to try to update.
	if le.IsLeader() {
		electionRecord.AcquireTime = oldElectionRecord.AcquireTime
		electionRecord.LeaderTransitions = oldElectionRecord.LeaderTransitions
	} else {
		electionRecord.LeaderTransitions = oldElectionRecord.LeaderTransitions + 1
	}
	if err = le.config.Lock.Update(ctx, electionRecord); err != nil {
		le.error("fail to update resource lock", err)
		return false
	}

	le.setObservedRecord(electionRecord)
	return true
}

func (le *LeaderElector) maybeReportTransition() {
	electionRecord, _ := le.getObservedRecord()
	if electionRecord.HolderIdentity == le.reportedLeader {
		return
	}

	le.reportedLeader = electionRecord.HolderIdentity
	if le.config.Callbacks.OnNewLeader != nil {
		go le.config.Callbacks.OnNewLeader(le.reportedLeader)
	}
}

func (le *LeaderElector) setObservedRecord(observedRecord ElectionRecord) {
	le.observedRecord.Store(observedElectionRecord{observedRecord, time.Now()})
}

func (le *LeaderElector) getObservedRecord() (ElectionRecord, time.Time) {
	rcd := le.observedRecord.Load().(observedElectionRecord)
	return rcd.ElectionRecord, rcd.Time
}

type observedElectionRecord struct {
	ElectionRecord
	Time time.Time
}
