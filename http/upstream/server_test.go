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

package upstream

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"testing"
	"time"
)

func TestServers(t *testing.T) {
	s1, _ := NewServer(ServerConfig{StaticWeight: 1, URL: URL{IP: "127.0.0.1", Port: 8001}})
	s2, _ := NewServer(ServerConfig{StaticWeight: 1, URL: URL{IP: "127.0.0.1", Port: 8002}})
	s3, _ := NewServer(ServerConfig{StaticWeight: 3, URL: URL{IP: "127.0.0.1", Port: 8003}})
	s4, _ := NewServer(ServerConfig{StaticWeight: 3, URL: URL{IP: "127.0.0.1", Port: 8004}})
	s5, _ := NewServer(ServerConfig{StaticWeight: 2, URL: URL{IP: "127.0.0.1", Port: 8005}})
	s6, _ := NewServer(ServerConfig{StaticWeight: 2, URL: URL{IP: "127.0.0.1", Port: 8006}})

	servers := Servers{s1, s2, s3, s4, s5, s6}
	sort.Stable(servers)

	exports := []uint16{8001, 8002, 8005, 8006, 8003, 8004}
	for i, server := range servers {
		if port := server.URL().Port; exports[i] != port {
			t.Errorf("expect the port '%d', but got '%d'", exports[i], port)
		}
	}
}

func ExampleURL_ID() {
	var url URL

	url = URL{Scheme: "http", IP: "127.0.0.1", Path: "/path"}
	fmt.Println(url.ID())

	url = URL{Scheme: "http", IP: "127.0.0.1", Port: 80, Path: "/path"}
	fmt.Println(url.ID())

	url = URL{Scheme: "http", Domain: "www.example.com", Path: "/path"}
	fmt.Println(url.ID())

	url = URL{Scheme: "http", Domain: "www.example.com", Port: 80, Path: "/path"}
	fmt.Println(url.ID())

	url = URL{Scheme: "http", Domain: "www.example.com", IP: "127.0.0.1", Path: "/path"}
	fmt.Println(url.ID())

	url = URL{Scheme: "http", Domain: "www.example.com", IP: "127.0.0.1", Port: 80, Path: "/path"}
	fmt.Println(url.ID())

	// Output:
	// http://127.0.0.1/path
	// http://127.0.0.1:80/path
	// http://www.example.com/path
	// http://www.example.com:80/path
	// http://www.example.com+127.0.0.1/path
	// http://www.example.com+127.0.0.1:80/path
}

func TestURL_Request(t *testing.T) {
	url := URL{IP: "127.0.0.1", Port: 8200}
	req, err := url.Request(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(201)
	})

	go http.ListenAndServe("127.0.0.1:8200", handler)
	time.Sleep(time.Millisecond * 100)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 201 {
		t.Errorf("expect status code '%d', but got '%d'", 201, resp.StatusCode)
	}
}

func TestServersPool(t *testing.T) {
	sp := NewServerPool(8)

	if servers := sp.Acquire(); cap(servers) != 8 {
		t.Errorf("expect %d servers, but got '%d'", 8, len(servers))
	}

	sp.Release(make(Servers, 10))
	if servers := sp.Acquire(); cap(servers) != 10 {
		t.Errorf("expect %d servers, but got '%d'", 10, len(servers))
	}
}
