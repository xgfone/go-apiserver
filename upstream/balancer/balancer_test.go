// Copyright 2021~2023 xgfone
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
	"testing"

	"github.com/xgfone/go-apiserver/nets"
	"github.com/xgfone/go-apiserver/upstream"
)

var _ upstream.Server = new(testServer)

type testServer struct {
	ip      string
	state   nets.RuntimeState
	weight  int
	current uint64
}

func (s *testServer) Weight() int                   { return s.weight }
func (s *testServer) ID() string                    { return s.ip }
func (s *testServer) Type() string                  { return "" }
func (s *testServer) Info() interface{}             { return nil }
func (s *testServer) Check(context.Context) error   { return nil }
func (s *testServer) Update(info interface{}) error { return nil }
func (s *testServer) Status() upstream.ServerStatus { return upstream.ServerStatusOnline }
func (s *testServer) RuntimeState() nets.RuntimeState {
	state := s.state.Clone()
	if s.current > 0 {
		state.Current = s.current
	}
	return state
}
func (s *testServer) Serve(c context.Context, r interface{}) error {
	s.state.Inc()
	s.state.Dec()
	return nil
}

func newTestServer(ip string, weight int) *testServer {
	return &testServer{ip: ip, weight: weight, current: uint64(weight)}
}

func TestRegisteredBuiltinBuidler(t *testing.T) {
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
		} else if _, err := Build(typ, nil); err != nil {
			t.Error(err)
		}
	}
}
