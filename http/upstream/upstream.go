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

// Package upstream provides the http request forwarder to the upstream servers.
package upstream

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sort"
	"sync"
	stdatomic "sync/atomic"
	"time"

	"github.com/xgfone/go-apiserver/internal/atomic"
	"github.com/xgfone/go-apiserver/log"
)

// ErrNoAvailableServers is used to represents no available servers.
var ErrNoAvailableServers = errors.New("no available servers")

// ServerInfo is the information of the upstream server.
type ServerInfo struct {
	failure int
	online  bool
	Server
}

// Online reports whether the server is online.
func (si ServerInfo) Online() bool { return si.online }

func (si *ServerInfo) updateStatus(online bool) (ok bool) {
	if ok = si.online != online; ok {
		si.online = online
	}
	return
}

type serversWrapper struct{ Servers }

// Upstream is used to route the request to one of a group of the backend servers.
type Upstream struct {
	name      string
	forwarder atomic.Value
	timeout   int64

	// Health Check
	hclock sync.RWMutex
	hcinfo healthCheckInfoWrapper

	slock   sync.RWMutex
	servers map[string]*ServerInfo
	server  atomic.Value // For serversWrapper
}

// NewUpstream returns a new upstream with the forwarder.
//
// If forwarder is nil, use Retry(RoundRobin()) by default.
//
// TODO: Add the retry when failed to forward the request.
func NewUpstream(name string, forwarder Forwarder) *Upstream {
	if forwarder == nil {
		forwarder = Retry(RoundRobin())
	}

	up := &Upstream{name: name, servers: make(map[string]*ServerInfo, 8)}
	up.server.Store(serversWrapper{})
	up.forwarder.Store(forwarder)
	return up
}

// Name reutrns the name of the upstream.
func (u *Upstream) Name() string { return u.name }

// GetForwarder returns the forwarder.
func (u *Upstream) GetForwarder() Forwarder {
	return u.forwarder.Load().(Forwarder)
}

// SwapForwarder swaps the old forwarder with the new.
func (u *Upstream) SwapForwarder(new Forwarder) (old Forwarder) {
	return u.forwarder.Swap(new).(Forwarder)
}

// GetTimeout returns the maximum timeout.
func (u *Upstream) GetTimeout() time.Duration {
	return time.Duration(stdatomic.LoadInt64(&u.timeout))
}

// SetTimeout sets the maximum timeout.
func (u *Upstream) SetTimeout(timeout time.Duration) {
	stdatomic.StoreInt64(&u.timeout, int64(timeout))
}

// HandleHTTP implements the interface Server.
func (u *Upstream) HandleHTTP(w http.ResponseWriter, r *http.Request) error {
	servers := u.server.Load().(serversWrapper).Servers
	if len(servers) == 0 {
		return ErrNoAvailableServers
	}

	if timeout := u.GetTimeout(); timeout > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		r = r.WithContext(ctx)
		defer cancel()
	}

	return u.GetForwarder().Forward(w, r, servers)
}

// ServeHTTP implements the interface http.Handler.
func (u *Upstream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := u.HandleHTTP(w, r)
	switch err {
	case nil:
		return

	case ErrNoAvailableServers:
		w.WriteHeader(503) // Service Unavailable

	default:
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			w.WriteHeader(504) // Gateway Timeout
		} else {
			w.WriteHeader(502) // Bad Gateway
		}
	}

	log.Error("fail to forward the http request",
		log.F("upstream", u.name),
		log.F("policy", u.GetForwarder().Policy()),
		log.F("clientaddr", r.RemoteAddr),
		log.F("reqhost", r.Host),
		log.F("reqmethod", r.Method),
		log.F("reqpath", r.URL.Path),
		log.E(err))
}

func (u *Upstream) updateServers() {
	servers := make(Servers, 0, len(u.servers))
	for _, si := range u.servers {
		if si.online {
			servers = append(servers, si.Server)
		}
	}
	sort.Stable(servers)
	u.server.Store(serversWrapper{servers})
}

// ResetServers resets all the servers.
func (u *Upstream) ResetServers(servers ...Server) {
	u.slock.Lock()
	defer u.slock.Unlock()

	u.servers = make(map[string]*ServerInfo, len(servers))
	for _len := len(servers) - 1; _len >= 0; _len-- {
		server := servers[_len]
		id := server.ID()
		u.servers[id] = &ServerInfo{Server: server, online: true}
	}
	u.updateServers()
}

// UpsertServers adds or updates the servers.
func (u *Upstream) UpsertServers(servers ...Server) {
	u.slock.Lock()
	defer u.slock.Unlock()

	for _len := len(servers) - 1; _len >= 0; _len-- {
		server := servers[_len]
		id := server.ID()
		if si, ok := u.servers[id]; ok {
			si.Server = server
			si.online = true
		} else {
			u.servers[id] = &ServerInfo{Server: server, online: true}
		}
	}
	u.updateServers()
}

// RemoveServer removes and returns the server by the server id.
//
// If the server does not exist, do nothing and return nil.
func (u *Upstream) RemoveServer(id string) (server Server) {
	u.slock.Lock()
	if si, ok := u.servers[id]; ok {
		delete(u.servers, id)
		u.updateServers()
		server = si.Server
	}
	u.slock.Unlock()
	return
}

// GetServer returns the server by the server id.
func (u *Upstream) GetServer(id string) (server ServerInfo, ok bool) {
	u.slock.RLock()
	si, ok := u.servers[id]
	if ok {
		server = *si
	}
	u.slock.RUnlock()
	return
}

// GetServers returns all the servers.
func (u *Upstream) GetServers() []ServerInfo {
	u.slock.RLock()
	servers := make([]ServerInfo, 0, len(u.servers))
	for _, si := range u.servers {
		servers = append(servers, *si)
	}
	u.slock.RUnlock()
	return servers
}

// SetServerStatus sets the online status of the server and return true.
// If the server does not exist, do nothing and return false.
func (u *Upstream) SetServerStatus(id string, online bool) (ok bool) {
	u.slock.Lock()
	ok = u.setServerStatus(id, online)
	u.slock.Unlock()
	return
}

func (u *Upstream) setServerStatus(id string, online bool) (ok bool) {
	si, ok := u.servers[id]
	if ok {
		si.updateStatus(online)
	}
	return
}

// SetServerStatuses sets the online status of the servers.
func (u *Upstream) SetServerStatuses(statuses map[string]bool) {
	u.slock.Lock()
	for id, online := range statuses {
		if si, ok := u.servers[id]; ok && si.online != online {
			si.online = online
		}
	}
	u.slock.Unlock()
}

/// ---------------------------------------------------------------------- ///

// HealthCheckInfo is the information of the health checker.
type HealthCheckInfo struct {
	URL

	Failure  int           `json:"failure" yaml:"failure"`
	Timeout  time.Duration `json:"timeout" yaml:"timeout"`
	Interval time.Duration `json:"interval" yaml:"interval"`
}

// IsZero reports whether the health check information is ZERO.
func (hci HealthCheckInfo) IsZero() bool {
	return hci.URL.IsZero() && hci.Failure == 0 && hci.Timeout == 0 && hci.Interval == 0
}

// Equal reports whether the health check information is equal to other.
func (hci HealthCheckInfo) Equal(other HealthCheckInfo) bool {
	return hci.Failure == other.Failure &&
		hci.Timeout == other.Timeout &&
		hci.Interval == other.Interval &&
		hci.URL.Equal(other.URL)
}

type healthCheckInfoWrapper struct {
	stop     chan struct{}
	ticker   *time.Ticker
	interval time.Duration
	HealthCheckInfo
}

// Close implements the interface io.Closer.
func (u *Upstream) Close() error {
	u.SetHealthCheck(HealthCheckInfo{})
	return nil
}

// GetHealthCheck returns the information of the health checker.
func (u *Upstream) GetHealthCheck() HealthCheckInfo {
	u.hclock.RLock()
	info := u.hcinfo.HealthCheckInfo
	u.hclock.RUnlock()
	return info
}

// SetHealthCheck resets the health check.
//
// If HealthCheckInfo is ZERO, it will clear the health check.
func (u *Upstream) SetHealthCheck(hci HealthCheckInfo) {
	u.hclock.Lock()
	defer u.hclock.Unlock()

	if hci.IsZero() {
		if u.hcinfo.stop != nil { // The health check is running.
			// Stop the health check and reset the health check to ZERO.
			close(u.hcinfo.stop)
			u.hcinfo.ticker.Stop()
			u.hcinfo = healthCheckInfoWrapper{}
		}
	} else if !u.hcinfo.HealthCheckInfo.Equal(hci) {
		interval := hci.Interval
		if interval <= 0 {
			interval = time.Minute
		}

		u.hcinfo.HealthCheckInfo = hci
		if u.hcinfo.interval != interval {
			u.hcinfo.interval = interval
			if u.hcinfo.stop == nil {
				u.hcinfo.stop = make(chan struct{})
				u.hcinfo.ticker = time.NewTicker(interval)
				go u.healthCheckLoop(u.hcinfo.stop, u.hcinfo.ticker)
			} else {
				u.hcinfo.ticker.Reset(interval)
			}
		}
	}
}

func (u *Upstream) healthCheckLoop(stop <-chan struct{}, ticker *time.Ticker) {
	for {
		select {
		case <-stop:
			return

		case <-ticker.C:
			u.hclock.RLock()
			hci := u.hcinfo.HealthCheckInfo
			u.hclock.RUnlock()
			u.checkServers(hci)
		}
	}
}

func (u *Upstream) checkServers(hci HealthCheckInfo) {
	u.slock.Lock()
	defer u.slock.Unlock()

	var changed bool
	for _, si := range u.servers {
		if online := u.checkServer(hci, si.Server); online {
			if si.updateStatus(online) && !changed {
				changed = true
			}

			if si.failure > 0 {
				si.failure = 0
			}
		} else if si.failure++; si.failure >= hci.Failure {
			if si.updateStatus(false) && !changed {
				changed = true
			}
		}
	}

	if changed {
		u.updateServers()
	}
}

func (u *Upstream) checkServer(hci HealthCheckInfo, server Server) (ok bool) {
	ctx := context.Background()
	if hci.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, hci.Timeout)
		defer cancel()
	}
	return server.Check(ctx, hci.URL) == nil
}
