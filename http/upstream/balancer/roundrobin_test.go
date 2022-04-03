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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/go-apiserver/http/upstream"
)

func TestWeightedRoundRobin(t *testing.T) {
	server1 := newTestServer("127.0.0.1", 1)
	server2 := newTestServer("127.0.0.2", 2)
	server3 := newTestServer("127.0.0.3", 3)
	servers := upstream.Servers{server1, server2, server3}

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)

	balancer := WeightedRoundRobin()
	for i := 0; i < 18; i++ {
		balancer.Forward(rec, req, servers)
	}

	if state := server1.State(); state.Total != 3 {
		t.Errorf("expect %d server1, but got %d", 3, state.Total)
	}
	if state := server2.State(); state.Total != 6 {
		t.Errorf("expect %d server2, but got %d", 6, state.Total)
	}
	if state := server3.State(); state.Total != 9 {
		t.Errorf("expect %d server3, but got %d", 9, state.Total)
	}
}
