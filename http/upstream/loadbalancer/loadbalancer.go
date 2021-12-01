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

// Package loadbalancer implements an upstream forwarder based on loadbalancer.
package loadbalancer

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sort"
	"sync"
	stdatomic "sync/atomic"
	"time"

	"github.com/xgfone/go-apiserver/http/upstream"
	"github.com/xgfone/go-apiserver/internal/atomic"
	"github.com/xgfone/go-apiserver/log"
)

// ErrNoAvailableServers is used to represents no available servers.
var ErrNoAvailableServers = errors.New("no available servers")

// ServerInfo is the information of the upstream server.
type ServerInfo struct {
	failure int
	online  bool
	upstream.Server
}

// Online reports whether the server is online.
func (si ServerInfo) Online() bool { return si.online }

func (si *ServerInfo) updateStatus(online bool) (ok bool) {
	if ok = si.online != online; ok {
		si.online = online
	}
	return
}

type serversWrapper struct{ upstream.Servers }

// LoadBalancer is used to forward the http request to one of the backend servers.
type LoadBalancer struct {
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

// NewLoadBalancer returns a new LoadBalancer to forward the http request.
//
// If forwarder is nil, use Retry(RoundRobin()) by default.
//
// TODO: Add the retry when failed to forward the request.
func NewLoadBalancer(name string, forwarder Forwarder) *LoadBalancer {
	if forwarder == nil {
		forwarder = Retry(RoundRobin())
	}

	up := &LoadBalancer{name: name, servers: make(map[string]*ServerInfo, 8)}
	up.server.Store(serversWrapper{})
	up.forwarder.Store(forwarder)
	return up
}

// Name reutrns the name of the upstream.
func (lb *LoadBalancer) Name() string { return lb.name }

// GetForwarder returns the forwarder.
func (lb *LoadBalancer) GetForwarder() Forwarder {
	return lb.forwarder.Load().(Forwarder)
}

// SwapForwarder swaps the old forwarder with the new.
func (lb *LoadBalancer) SwapForwarder(new Forwarder) (old Forwarder) {
	return lb.forwarder.Swap(new).(Forwarder)
}

// GetTimeout returns the maximum timeout.
func (lb *LoadBalancer) GetTimeout() time.Duration {
	return time.Duration(stdatomic.LoadInt64(&lb.timeout))
}

// SetTimeout sets the maximum timeout.
func (lb *LoadBalancer) SetTimeout(timeout time.Duration) {
	stdatomic.StoreInt64(&lb.timeout, int64(timeout))
}

// HandleHTTP implements the interface Server.
func (lb *LoadBalancer) HandleHTTP(w http.ResponseWriter, r *http.Request) error {
	servers := lb.server.Load().(serversWrapper).Servers
	if len(servers) == 0 {
		return ErrNoAvailableServers
	}

	if timeout := lb.GetTimeout(); timeout > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		r = r.WithContext(ctx)
		defer cancel()
	}

	return lb.GetForwarder().Forward(w, r, servers)
}

// ServeHTTP implements the interface http.Handler.
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := lb.HandleHTTP(w, r)
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
		log.F("upstream", lb.name),
		log.F("policy", lb.GetForwarder().Policy()),
		log.F("clientaddr", r.RemoteAddr),
		log.F("reqhost", r.Host),
		log.F("reqmethod", r.Method),
		log.F("reqpath", r.URL.Path),
		log.E(err))
}

func (lb *LoadBalancer) updateServers() {
	servers := make(upstream.Servers, 0, len(lb.servers))
	for _, si := range lb.servers {
		if si.online {
			servers = append(servers, si.Server)
		}
	}
	sort.Stable(servers)
	lb.server.Store(serversWrapper{servers})
}

// ResetServers resets all the servers.
func (lb *LoadBalancer) ResetServers(servers ...upstream.Server) {
	lb.slock.Lock()
	defer lb.slock.Unlock()

	lb.servers = make(map[string]*ServerInfo, len(servers))
	for _len := len(servers) - 1; _len >= 0; _len-- {
		server := servers[_len]
		id := server.ID()
		lb.servers[id] = &ServerInfo{Server: server, online: true}
	}
	lb.updateServers()
}

// UpsertServers adds or updates the servers.
func (lb *LoadBalancer) UpsertServers(servers ...upstream.Server) {
	lb.slock.Lock()
	defer lb.slock.Unlock()

	for _len := len(servers) - 1; _len >= 0; _len-- {
		server := servers[_len]
		id := server.ID()
		if si, ok := lb.servers[id]; ok {
			si.Server = server
			si.online = true
		} else {
			lb.servers[id] = &ServerInfo{Server: server, online: true}
		}
	}
	lb.updateServers()
}

// RemoveServer removes and returns the server by the server id.
//
// If the server does not exist, do nothing and return nil.
func (lb *LoadBalancer) RemoveServer(id string) (server upstream.Server) {
	lb.slock.Lock()
	if si, ok := lb.servers[id]; ok {
		delete(lb.servers, id)
		lb.updateServers()
		server = si.Server
	}
	lb.slock.Unlock()
	return
}

// GetServer returns the server by the server id.
func (lb *LoadBalancer) GetServer(id string) (server ServerInfo, ok bool) {
	lb.slock.RLock()
	si, ok := lb.servers[id]
	if ok {
		server = *si
	}
	lb.slock.RUnlock()
	return
}

// GetServers returns all the servers.
func (lb *LoadBalancer) GetServers() []ServerInfo {
	lb.slock.RLock()
	servers := make([]ServerInfo, 0, len(lb.servers))
	for _, si := range lb.servers {
		servers = append(servers, *si)
	}
	lb.slock.RUnlock()
	return servers
}

// SetServerStatus sets the online status of the server and return true.
// If the server does not exist, do nothing and return false.
func (lb *LoadBalancer) SetServerStatus(id string, online bool) (ok bool) {
	lb.slock.Lock()
	ok = lb.setServerStatus(id, online)
	lb.slock.Unlock()
	return
}

func (lb *LoadBalancer) setServerStatus(id string, online bool) (ok bool) {
	si, ok := lb.servers[id]
	if ok {
		si.updateStatus(online)
	}
	return
}

// SetServerStatuses sets the online status of the servers.
func (lb *LoadBalancer) SetServerStatuses(statuses map[string]bool) {
	lb.slock.Lock()
	for id, online := range statuses {
		if si, ok := lb.servers[id]; ok && si.online != online {
			si.online = online
		}
	}
	lb.slock.Unlock()
}

/// ---------------------------------------------------------------------- ///

// HealthCheckInfo is the information of the health checker.
type HealthCheckInfo struct {
	upstream.URL

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
func (lb *LoadBalancer) Close() error {
	lb.SetHealthCheck(HealthCheckInfo{})
	return nil
}

// GetHealthCheck returns the information of the health checker.
func (lb *LoadBalancer) GetHealthCheck() HealthCheckInfo {
	lb.hclock.RLock()
	info := lb.hcinfo.HealthCheckInfo
	lb.hclock.RUnlock()
	return info
}

// SetHealthCheck resets the health check.
//
// If HealthCheckInfo is ZERO, it will clear the health check.
func (lb *LoadBalancer) SetHealthCheck(hci HealthCheckInfo) {
	lb.hclock.Lock()
	defer lb.hclock.Unlock()

	if hci.IsZero() {
		if lb.hcinfo.stop != nil { // The health check is running.
			// Stop the health check and reset the health check to ZERO.
			close(lb.hcinfo.stop)
			lb.hcinfo.ticker.Stop()
			lb.hcinfo = healthCheckInfoWrapper{}
		}
	} else if !lb.hcinfo.HealthCheckInfo.Equal(hci) {
		interval := hci.Interval
		if interval <= 0 {
			interval = time.Minute
		}

		lb.hcinfo.HealthCheckInfo = hci
		if lb.hcinfo.interval != interval {
			lb.hcinfo.interval = interval
			if lb.hcinfo.stop == nil {
				lb.hcinfo.stop = make(chan struct{})
				lb.hcinfo.ticker = time.NewTicker(interval)
				go lb.healthCheckLoop(lb.hcinfo.stop, lb.hcinfo.ticker)
			} else {
				lb.hcinfo.ticker.Reset(interval)
			}
		}
	}
}

func (lb *LoadBalancer) healthCheckLoop(stop <-chan struct{}, ticker *time.Ticker) {
	for {
		select {
		case <-stop:
			return

		case <-ticker.C:
			lb.hclock.RLock()
			hci := lb.hcinfo.HealthCheckInfo
			lb.hclock.RUnlock()
			lb.checkServers(hci)
		}
	}
}

func (lb *LoadBalancer) checkServers(hci HealthCheckInfo) {
	lb.slock.Lock()
	defer lb.slock.Unlock()

	var changed bool
	for _, si := range lb.servers {
		if online := lb.checkServer(hci, si.Server); online {
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
		lb.updateServers()
	}
}

func (lb *LoadBalancer) checkServer(hci HealthCheckInfo, server upstream.Server) (ok bool) {
	ctx := context.Background()
	if hci.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, hci.Timeout)
		defer cancel()
	}
	return server.Check(ctx, hci.URL) == nil
}
