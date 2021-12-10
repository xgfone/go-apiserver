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

package balancer

import (
	"context"
	"net/http"
	"testing"

	"github.com/xgfone/go-apiserver/http/upstream"
	"github.com/xgfone/go-apiserver/nets"
)

type testServer struct {
	url     upstream.URL
	state   nets.RuntimeState
	weight  int
	current uint64
}

func (s *testServer) Weight() int                               { return s.weight }
func (s *testServer) ID() string                                { return s.url.IP }
func (s *testServer) URL() upstream.URL                         { return s.url }
func (s *testServer) Check(context.Context, upstream.URL) error { return nil }
func (s *testServer) State() nets.RuntimeState {
	state := s.state.Clone()
	if s.current > 0 {
		state.Current = s.current
	}
	return state
}
func (s *testServer) HandleHTTP(http.ResponseWriter, *http.Request) error {
	s.state.Inc()
	s.state.Dec()
	return nil
}

func newTestServer(ip string, weight int) *testServer {
	return &testServer{
		url:     upstream.URL{IP: ip},
		weight:  weight,
		current: uint64(weight),
	}
}

func TestRegisterBuidler(t *testing.T) {
	var callback1 SelectedServerCallback
	callback2 := func(*http.Request, upstream.Server) {}

	expects := []string{
		"random",
		"round_robin",
		"weight_random",
		"weight_round_robin",
		"source_ip_hash",
		"least_conn",
	}

	for _, typ := range expects {
		if _, err := Build(typ, nil); err != nil {
			t.Error(err)
		} else if _, err := Build(typ, callback1); err != nil {
			t.Error(err)
		} else if _, err := Build(typ, callback2); err != nil {
			t.Error(err)
		}
	}
}

func TestCalcServerOnWeight(t *testing.T) {
	server1, _ := upstream.NewServer(upstream.ServerConfig{
		URL: upstream.URL{IP: "127.0.0.1", Port: 8101}, StaticWeight: 1,
	})
	server2, _ := upstream.NewServer(upstream.ServerConfig{
		URL: upstream.URL{IP: "127.0.0.1", Port: 8101}, StaticWeight: 2,
	})
	server3, _ := upstream.NewServer(upstream.ServerConfig{
		URL: upstream.URL{IP: "127.0.0.1", Port: 8101}, StaticWeight: 3,
	})
	servers := upstream.Servers{server1, server2, server3}

	if server := calcServerOnWeight(servers, 1); server.ID() != server3.ID() {
		t.Errorf("expect server '%s', but got '%s'", server3.ID(), server.ID())
	}
	if server := calcServerOnWeight(servers, 2); server.ID() != server3.ID() {
		t.Errorf("expect server '%s', but got '%s'", server3.ID(), server.ID())
	}
	if server := calcServerOnWeight(servers, 3); server.ID() != server3.ID() {
		t.Errorf("expect server '%s', but got '%s'", server3.ID(), server.ID())
	}

	if server := calcServerOnWeight(servers, 4); server.ID() != server2.ID() {
		t.Errorf("expect server '%s', but got '%s'", server2.ID(), server.ID())
	}
	if server := calcServerOnWeight(servers, 5); server.ID() != server2.ID() {
		t.Errorf("expect server '%s', but got '%s'", server2.ID(), server.ID())
	}

	if server := calcServerOnWeight(servers, 6); server.ID() != server1.ID() {
		t.Errorf("expect server '%s', but got '%s'", server1.ID(), server.ID())
	}

	if server := calcServerOnWeight(servers, 7); server.ID() != server3.ID() {
		t.Errorf("expect server '%s', but got '%s'", server3.ID(), server.ID())
	}
	if server := calcServerOnWeight(servers, 8); server.ID() != server3.ID() {
		t.Errorf("expect server '%s', but got '%s'", server3.ID(), server.ID())
	}
	if server := calcServerOnWeight(servers, 9); server.ID() != server3.ID() {
		t.Errorf("expect server '%s', but got '%s'", server3.ID(), server.ID())
	}

	if server := calcServerOnWeight(servers, 10); server.ID() != server2.ID() {
		t.Errorf("expect server '%s', but got '%s'", server2.ID(), server.ID())
	}
}
