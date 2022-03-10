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

	url = URL{Scheme: "http", Hostname: "www.example.com", Path: "/path"}
	fmt.Println(url.ID())

	url = URL{Scheme: "http", Hostname: "www.example.com", Port: 80, Path: "/path"}
	fmt.Println(url.ID())

	url = URL{Scheme: "http", Hostname: "www.example.com", IP: "127.0.0.1", Path: "/path"}
	fmt.Println(url.ID())

	url = URL{Scheme: "http", Hostname: "www.example.com", IP: "127.0.0.1", Port: 80, Path: "/path"}
	fmt.Println(url.ID())

	// Output:
	// http://127.0.0.1/path#md5=21aca36be0bd34307f635553a460db41
	// http://127.0.0.1:80/path#md5=3da30ab9783aad141993ce4e2940608a
	// http://www.example.com/path#md5=8aa32ab56942b28249eaf6e06ecb3d08
	// http://www.example.com:80/path#md5=1c622fa8baecdf9570ecb95e89249f02
	// http://www.example.com+127.0.0.1/path#md5=32243ff8dfc9ac922946dcd0a89cc1b9
	// http://www.example.com+127.0.0.1:80/path#md5=b4729cc202e4b573fd33563c4c496adc
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
