// Copyright 2021 ~2023xgfone
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

	"github.com/xgfone/go-apiserver/upstream"
)

func TestSourceIPHash(t *testing.T) {
	server1 := newTestServer("127.0.0.1", 1)
	server2 := newTestServer("127.0.0.2", 2)
	server3 := newTestServer("127.0.0.3", 3)
	servers := upstream.Servers{server1, server2, server3}

	req1, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	req2, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	req3, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	req1.RemoteAddr = "192.168.0.0"
	req2.RemoteAddr = "192.168.0.1"
	req3.RemoteAddr = "192.168.0.2"

	balancer := SourceIPHash()
	balancer.Forward(context.Background(), req1, servers)
	balancer.Forward(context.Background(), req1, servers)
	balancer.Forward(context.Background(), req1, servers)
	balancer.Forward(context.Background(), req1, servers)
	balancer.Forward(context.Background(), req1, servers)
	balancer.Forward(context.Background(), req1, servers)
	balancer.Forward(context.Background(), req2, servers)
	balancer.Forward(context.Background(), req3, servers)

	if total := server1.RuntimeState().Total; total != 6 {
		t.Errorf("expect %d server1, but got %d", 6, total)
	}
	if total := server2.RuntimeState().Total; total != 1 {
		t.Errorf("expect %d server1, but got %d", 1, total)
	}
	if total := server3.RuntimeState().Total; total != 1 {
		t.Errorf("expect %d server1, but got %d", 1, total)
	}
}