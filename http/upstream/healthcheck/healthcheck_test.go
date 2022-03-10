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

package healthcheck

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/xgfone/go-apiserver/http/upstream"
	"github.com/xgfone/go-apiserver/nets"
)

type testServer struct{ id string }

func newServer(id string) testServer { return testServer{id: id} }

func (s testServer) ID() string                                          { return s.id }
func (s testServer) URL() (url upstream.URL)                             { return }
func (s testServer) State() (rs nets.RuntimeState)                       { return }
func (s testServer) HandleHTTP(http.ResponseWriter, *http.Request) error { return nil }
func (s testServer) Check(context.Context, upstream.URL) error {
	return nil
}

type testUpdater struct{ servers sync.Map }

func newUpdater() *testUpdater { return &testUpdater{} }

func (u *testUpdater) UpsertServer(s upstream.Server)         { u.servers.Store(s.ID(), false) }
func (u *testUpdater) RemoveServer(id string)                 { u.servers.Delete(id) }
func (u *testUpdater) SetServerOnline(id string, online bool) { u.servers.Store(id, online) }
func (u *testUpdater) Servers() map[string]bool {
	servers := make(map[string]bool)
	u.servers.Range(func(key, value interface{}) bool {
		servers[key.(string)] = value.(bool)
		return true
	})
	return servers
}

func TestHealthCheck(t *testing.T) {
	updater1 := newUpdater()
	updater2 := newUpdater()

	hc := NewHealthChecker(time.Millisecond * 100)
	hc.Start()
	defer hc.Stop()

	hc.AddUpdater("updater1", updater1)
	hc.UpsertServer(newServer("id1"), Info{})
	hc.UpsertServer(newServer("id2"), Info{})
	hc.AddUpdater("updater2", updater2)

	time.Sleep(time.Millisecond * 500)

	servers := make(map[string]bool)
	for _, sc := range hc.GetServers() {
		servers[sc.Server.ID()] = sc.Online
	}

	checkServers(t, "hc", servers)
	checkServers(t, "updater1", updater1.Servers())
	checkServers(t, "updater2", updater2.Servers())
}

func checkServers(t *testing.T, prefix string, servers map[string]bool) {
	if len(servers) != 2 {
		t.Errorf("%s: expect %d servers, but got %d", prefix, 2, len(servers))
	} else {
		for sid, online := range servers {
			switch sid {
			case "id1", "id2":
			default:
				t.Errorf("%s: unexpected server '%s'", prefix, sid)
			}

			if !online {
				t.Errorf("%s: the server '%s' is not online", prefix, sid)
			}
		}
	}
}
