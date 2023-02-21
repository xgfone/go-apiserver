// Copyright 2022~2023 xgfone
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

	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/upstream"
)

var (
	// DefaultHealthChecker is the default global health checker.
	DefaultHealthChecker = NewHealthChecker()

	// DefaultInterval is the default healthcheck interval.
	DefaultInterval = time.Second * 10

	// DefaultCheckConfig is the default health check configuration.
	DefaultCheckConfig = CheckConfig{Failure: 1, Timeout: time.Second, Interval: DefaultInterval}
)

// CheckConfig is the config to check the server health.
type CheckConfig struct {
	Failure  int           `json:"failure,omitempty"`
	Timeout  time.Duration `json:"timeout,omitempty"`
	Interval time.Duration `json:"interval,omitempty"`
	Delay    time.Duration `json:"delay,omitempty"`
}

type serverWrapper struct{ upstream.Server }

type serverContext struct {
	tickch  chan time.Duration
	stopch  chan struct{}
	config  atomic.Value
	server  atomic.Value
	online  int32
	failure int
}

func newServerContext(server upstream.Server, config CheckConfig) *serverContext {
	sc := &serverContext{tickch: make(chan time.Duration), stopch: make(chan struct{})}
	sc.SetServer(server)
	sc.SetConfig(config)
	return sc
}

// IsOnline reports whether the server is online.
func (s *serverContext) IsOnline() bool {
	return atomic.LoadInt32(&s.online) == 1
}

func (s *serverContext) SetConfig(config CheckConfig) {
	if config.Interval <= 0 {
		if DefaultInterval > 0 {
			config.Interval = DefaultInterval
		} else {
			config.Interval = time.Second * 10
		}
	}

	s.config.Store(config)
	select {
	case s.tickch <- config.Interval:
	default:
	}
}

func (s *serverContext) GetConfig() CheckConfig {
	return s.config.Load().(CheckConfig)
}

func (s *serverContext) GetServer() upstream.Server {
	return s.server.Load().(serverWrapper).Server
}

func (s *serverContext) SetServer(server upstream.Server) {
	s.server.Store(serverWrapper{server})
}

func (s *serverContext) Stop() {
	close(s.stopch)
}

func (s *serverContext) beforeStart(hc *HealthChecker, exit <-chan struct{}) {
	config := s.GetConfig()
	if config.Delay > 0 {
		wait := time.NewTimer(config.Delay)
		select {
		case <-wait.C:
		case <-exit:
			wait.Stop()
			return
		}
	}
	s.checkHealth(hc, config)
}

func (s *serverContext) Start(hc *HealthChecker, exit <-chan struct{}) {
	s.beforeStart(hc, exit)
	ticker := time.NewTicker(s.GetConfig().Interval)
	defer ticker.Stop()

	for {
		select {
		case <-exit:
			return

		case <-s.stopch:
			return

		case tick := <-s.tickch:
			ticker.Reset(tick)

		case <-ticker.C:
			s.checkHealth(hc, s.GetConfig())
		}
	}
}

func (s *serverContext) wrapPanic() {
	if err := recover(); err != nil {
		log.Error("wrap a panic when to check the server health",
			"serverid", s.GetServer().ID())
	}
}

func (s *serverContext) checkHealth(hc *HealthChecker, config CheckConfig) {
	defer s.wrapPanic()
	if s.updateOnlineStatus(s.checkServer(config), config.Failure) {
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

func (s *serverContext) checkServer(config CheckConfig) (online bool) {
	ctx := context.Background()
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	server := s.GetServer()
	online = server.Check(ctx) == nil
	log.Debug("HealthChecker: check the upstream server", "serverid", server.ID(), "online", online)
	return
}

// HealthChecker is a health checker to check whether a set of http servers
// are healthy.
//
// Notice: if there are lots of servers to be checked, you maybe need
// an external checker.
type HealthChecker struct {
	exit     chan struct{}
	slock    sync.RWMutex
	servers  map[string]*serverContext
	updaters sync.Map
}

// NewHealthChecker returns a new health checker.
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{servers: make(map[string]*serverContext, 16)}
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
	defer hc.slock.Unlock()
	if hc.exit != nil {
		close(hc.exit)
		hc.exit = nil
	}
}

// Start starts the health checker.
func (hc *HealthChecker) Start() {
	hc.slock.Lock()
	defer hc.slock.Unlock()

	if hc.exit == nil {
		hc.exit = make(chan struct{})
		for _, sc := range hc.servers {
			go sc.Start(hc, hc.exit)
		}
	} else {
		panic("HealthChecker: has been started")
	}
}

// AddUpdater adds the healthcheck updater with the name.
func (hc *HealthChecker) AddUpdater(name string, updater Updater) (err error) {
	if name == "" {
		panic("HealthChecker: the name is empty")
	} else if updater == nil {
		panic("HealthChecker: the updater is nil")
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

// UpsertServers adds or updates a set of the servers with the same healthcheck config.
func (hc *HealthChecker) UpsertServers(servers upstream.Servers, config CheckConfig) {
	hc.slock.Lock()
	defer hc.slock.Unlock()
	for _, server := range servers {
		hc.upsertServer(server, config)
	}
}

// UpsertServer adds or updates the server with the healthcheck config.
func (hc *HealthChecker) UpsertServer(server upstream.Server, config CheckConfig) {
	hc.slock.Lock()
	defer hc.slock.Unlock()
	hc.upsertServer(server, config)
}

func (hc *HealthChecker) upsertServer(server upstream.Server, config CheckConfig) {
	id := server.ID()
	sc, ok := hc.servers[id]
	if ok {
		sc.SetServer(server)
		sc.SetConfig(config)
	} else {
		sc = newServerContext(server, config)
		hc.servers[id] = sc
		if hc.exit != nil { // Has started
			go sc.Start(hc, hc.exit)
		}

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
		server.Config = sc.GetConfig()
	}
	return
}

var (
	_ upstream.Server        = ServerInfo{}
	_ upstream.ServerWrapper = ServerInfo{}
)

// ServerInfo represents the information of the server.
type ServerInfo struct {
	upstream.Server
	Config CheckConfig
	Online bool
}

// Unwrap unwraps the inner upstream server.
func (si ServerInfo) Unwrap() upstream.Server {
	return si.Server
}

// Status overrides the interface method upstream.Server#Status.
func (si ServerInfo) Status() upstream.ServerStatus {
	if si.Online {
		return upstream.ServerStatusOnline
	}
	return upstream.ServerStatusOffline
}

// GetServers returns all the servers.
func (hc *HealthChecker) GetServers() []ServerInfo {
	hc.slock.RLock()
	servers := make([]ServerInfo, 0, len(hc.servers))
	for _, sc := range hc.servers {
		servers = append(servers, ServerInfo{
			Config: sc.GetConfig(),
			Server: sc.GetServer(),
			Online: sc.IsOnline(),
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
