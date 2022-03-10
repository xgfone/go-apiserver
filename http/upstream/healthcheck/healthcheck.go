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

// Package healthcheck provides a health checker to check whether a set
// of http servers are healthy.
package healthcheck

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xgfone/go-apiserver/http/upstream"
	"github.com/xgfone/go-apiserver/log"
)

// Updater is used to update the server status.
type Updater interface {
	UpsertServer(upstream.Server)
	RemoveServer(serverID string)
	SetServerOnline(serverID string, online bool)
}

// DefaultHealthCheckInfo is the default healthcheck information.
var DefaultHealthCheckInfo = Info{Failure: 1, Timeout: time.Second, Interval: time.Second * 10}

// Info is the information of the health check.
type Info struct {
	upstream.URL `json:"url" yaml:"url"`

	Failure  int           `json:"failure,omitempty" yaml:"failure,omitempty"`
	Timeout  time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Interval time.Duration `json:"interval,omitempty" yaml:"interval,omitempty"`
}

// IsZero reports whether the health check information is ZERO.
func (i Info) IsZero() bool {
	return i.URL.IsZero() && i.Failure == 0 && i.Timeout == 0 && i.Interval == 0
}

// Equal reports whether the health check information is equal to other.
func (i Info) Equal(other Info) bool {
	return i.Failure == other.Failure &&
		i.Timeout == other.Timeout &&
		i.Interval == other.Interval &&
		i.URL.Equal(other.URL)
}

type serverWrapper struct{ upstream.Server }

type serverContext struct {
	stop   chan struct{}
	info   atomic.Value
	server atomic.Value
	online int32

	failure  int
	lasttime time.Time
}

func newServerContext(server upstream.Server, info Info) *serverContext {
	sc := &serverContext{}
	sc.SetServer(server)
	sc.SetInfo(info)
	return sc
}

// IsOnline reports whether the server is online.
func (s *serverContext) IsOnline() bool {
	return atomic.LoadInt32(&s.online) == 1
}

func (s *serverContext) SetInfo(info Info) { s.info.Store(info) }

func (s *serverContext) GetInfo() Info { return s.info.Load().(Info) }

func (s *serverContext) GetServer() upstream.Server {
	return s.server.Load().(serverWrapper).Server
}

func (s *serverContext) SetServer(server upstream.Server) {
	s.server.Store(serverWrapper{server})
}

func (s *serverContext) Stop() { close(s.stop) }

func (s *serverContext) Start(hc *HealthChecker) {
	exit := hc.exit
	s.checkHealth(hc, s.GetInfo())
	s.stop = make(chan struct{})

	tick := time.NewTicker(hc.tick)
	defer tick.Stop()

	for {
		select {
		case <-exit:
			return

		case <-s.stop:
			return

		case now := <-tick.C:
			info := s.GetInfo()
			if now.Sub(s.lasttime) > info.Interval {
				s.checkHealth(hc, info)
			}
		}
	}
}

func (s *serverContext) wrapPanic() {
	if err := recover(); err != nil {
		log.Error("wrap a panic when to check the server health",
			"serverid", s.GetServer().ID())
	}
}

func (s *serverContext) checkHealth(hc *HealthChecker, info Info) {
	defer s.wrapPanic()
	if s.updateOnlineStatus(s.checkServer(info), info.Failure) {
		hc.setOnline(s.GetServer().ID(), s.IsOnline())
	}
}

func (s *serverContext) updateOnlineStatus(online bool, failure int) (ok bool) {
	if online {
		if s.failure > 0 {
			s.failure = 0
		}
		ok = atomic.CompareAndSwapInt32(&s.online, 0, 1)
	} else if s.failure++; s.failure > failure {
		ok = atomic.CompareAndSwapInt32(&s.online, 1, 0)
	}
	return
}

func (s *serverContext) checkServer(info Info) (online bool) {
	ctx := context.Background()
	if info.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, info.Timeout)
		defer cancel()
	}
	return s.GetServer().Check(ctx, info.URL) == nil
}

// HealthChecker is a health checker to check whether a set of http servers
// are healthy.
type HealthChecker struct {
	tick     time.Duration
	exit     chan struct{}
	slock    sync.RWMutex
	servers  map[string]*serverContext
	updaters sync.Map
}

// NewHealthChecker returns a new health checker with the tick duration.
func NewHealthChecker(tick time.Duration) *HealthChecker {
	return &HealthChecker{
		tick:    tick,
		servers: make(map[string]*serverContext, 16),
	}
}

func (hc *HealthChecker) setOnline(serverID string, online bool) {
	hc.updaters.Range(func(_, value interface{}) bool {
		value.(Updater).SetServerOnline(serverID, online)
		return true
	})
}

// Stop stops the health checker.
func (hc *HealthChecker) Stop() {
	hc.slock.Lock()
	if hc.exit != nil {
		close(hc.exit)
		hc.exit = nil
	}
	hc.slock.Unlock()
}

// Start starts the health checker.
func (hc *HealthChecker) Start() {
	hc.slock.Lock()
	if hc.exit == nil {
		hc.exit = make(chan struct{})
		for _, sc := range hc.servers {
			go sc.Start(hc)
		}
	}
	hc.slock.Unlock()
}

// AddUpdater adds the healthcheck updater with the name.
func (hc *HealthChecker) AddUpdater(name string, updater Updater) (err error) {
	if name == "" {
		panic("the healthcheck name is empty")
	} else if updater == nil {
		panic("the healthcheck is nil")
	}

	if _, loaded := hc.updaters.LoadOrStore(name, updater); loaded {
		err = fmt.Errorf("the healthcheck updater named '%s' has been added", name)
	} else {
		hc.slock.RLock()
		for id, sc := range hc.servers {
			updater.UpsertServer(sc.GetServer())
			updater.SetServerOnline(id, sc.IsOnline())
		}
		hc.slock.RUnlock()
	}
	return
}

// DelUpdater deletes the healthcheck updater by the name.
func (hc *HealthChecker) DelUpdater(name string) {
	if name == "" {
		panic("the healthcheck name is empty")
	}
	hc.updaters.Delete(name)
}

// GetUpdater returns the healthcheck updater by the name.
//
// Return nil if the healthcheck updater does not exist.
func (hc *HealthChecker) GetUpdater(name string) Updater {
	if name == "" {
		panic("the healthcheck name is empty")
	}

	if value, ok := hc.updaters.Load(name); ok {
		return value.(Updater)
	}
	return nil
}

// GetUpdaters returns all the healthcheck updaters.
func (hc *HealthChecker) GetUpdaters() map[string]Updater {
	updaters := make(map[string]Updater, 32)
	hc.updaters.Range(func(key, value interface{}) bool {
		updaters[key.(string)] = value.(Updater)
		return true
	})
	return updaters
}

// UpsertServer adds or updates the server with the healthcheck information.
func (hc *HealthChecker) UpsertServer(server upstream.Server, healthCheck Info) {
	id := server.ID()

	hc.slock.Lock()
	sc, ok := hc.servers[id]
	if ok {
		sc.SetServer(server)
		sc.SetInfo(healthCheck)
	} else {
		sc = newServerContext(server, healthCheck)
		hc.servers[id] = sc

		if hc.exit != nil { // Has started
			go sc.Start(hc)
		}
	}
	hc.slock.Unlock()

	if !ok {
		hc.updaters.Range(func(_, value interface{}) bool {
			updater := value.(Updater)
			updater.UpsertServer(server)
			updater.SetServerOnline(id, sc.IsOnline())
			return true
		})
	}
}

// RemoveServer removes the server by the server id.
func (hc *HealthChecker) RemoveServer(serverID string) {
	if serverID == "" {
		panic("the upstream server id is empty")
	}

	hc.slock.Lock()
	if sc, ok := hc.servers[serverID]; ok {
		delete(hc.servers, serverID)
		hc.slock.Unlock()

		hc.updaters.Range(func(_, value interface{}) bool {
			value.(Updater).RemoveServer(serverID)
			return true
		})

		sc.Stop()
	} else {
		hc.slock.Unlock()
	}
}

// GetServer returns the server by the server id.
//
// Return nil if the server does not exist.
func (hc *HealthChecker) GetServer(serverID string) (server ServerInfo, ok bool) {
	if serverID == "" {
		panic("the upstream server id is empty")
	}

	hc.slock.RLock()
	sc, ok := hc.servers[serverID]
	hc.slock.RUnlock()

	if ok {
		server.Online = sc.IsOnline()
		server.Server = sc.GetServer()
		server.HealthCheck = sc.GetInfo()
	}
	return
}

// ServerInfo represents the information of the server.
type ServerInfo struct {
	HealthCheck Info

	Server upstream.Server
	Online bool
}

// GetServers returns all the servers.
func (hc *HealthChecker) GetServers() []ServerInfo {
	hc.slock.RLock()
	servers := make([]ServerInfo, 0, len(hc.servers))
	for _, sc := range hc.servers {
		servers = append(servers, ServerInfo{
			HealthCheck: sc.GetInfo(),
			Server:      sc.GetServer(),
			Online:      sc.IsOnline(),
		})
	}
	hc.slock.RUnlock()
	return servers
}

// ServerIsOnline reports whether the server is online.
func (hc *HealthChecker) ServerIsOnline(serverID string) (online, ok bool) {
	if serverID == "" {
		panic("the upstream server id is empty")
	}

	hc.slock.RLock()
	sc, ok := hc.servers[serverID]
	hc.slock.RUnlock()

	if ok {
		online = sc.IsOnline()
	}
	return
}
