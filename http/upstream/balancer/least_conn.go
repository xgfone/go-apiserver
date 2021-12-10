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
	"sort"

	"github.com/xgfone/go-apiserver/http/upstream"
)

func init() {
	registerBuiltinBuidler("least_conn", LeastConn)
}

// LeastConn returns a new balancer based on the least number of the connection.
//
// The policy name is "least_conn".
func LeastConn(callback SelectedServerCallback) Balancer {
	return NewForwarder("least_conn",
		func(w http.ResponseWriter, r *http.Request, ss upstream.Servers) (err error) {
			servers := upstream.DefaultServersPool.Acquire()
			servers = append(servers, ss...)
			sort.Stable(leastConnServers(servers))
			err = serverCallback(callback, w, r, servers[0])
			upstream.DefaultServersPool.Release(servers)
			return err
		})
}

type leastConnServers upstream.Servers

func (ss leastConnServers) Len() int      { return len(ss) }
func (ss leastConnServers) Swap(i, j int) { ss[i], ss[j] = ss[j], ss[i] }
func (ss leastConnServers) Less(i, j int) bool {
	return ss[i].State().Current < ss[j].State().Current
}
