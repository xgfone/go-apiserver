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

package httpserver

import (
	"sort"
	"testing"

	"github.com/xgfone/go-apiserver/upstream"
)

func TestServers(t *testing.T) {
	s1, _ := Config{StaticWeight: 1, URL: URL{IP: "127.0.0.1", Port: 8001}}.NewServer()
	s2, _ := Config{StaticWeight: 1, URL: URL{IP: "127.0.0.1", Port: 8002}}.NewServer()
	s3, _ := Config{StaticWeight: 3, URL: URL{IP: "127.0.0.1", Port: 8003}}.NewServer()
	s4, _ := Config{StaticWeight: 3, URL: URL{IP: "127.0.0.1", Port: 8004}}.NewServer()
	s5, _ := Config{StaticWeight: 2, URL: URL{IP: "127.0.0.1", Port: 8005}}.NewServer()
	s6, _ := Config{StaticWeight: 2, URL: URL{IP: "127.0.0.1", Port: 8006}}.NewServer()

	servers := upstream.Servers{s1, s2, s3, s4, s5, s6}
	sort.Stable(servers)

	exports := []uint16{8001, 8002, 8005, 8006, 8003, 8004}
	for i, server := range servers {
		if port := server.Info().(Config).URL.Port; exports[i] != port {
			t.Errorf("expect the port '%d', but got '%d'", exports[i], port)
		}
	}
}
