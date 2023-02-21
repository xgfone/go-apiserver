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
	"encoding/binary"
	"fmt"
	"net"
	"net/http"

	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/nets"
	"github.com/xgfone/go-apiserver/upstream"
)

func init() {
	registerBuiltinBuidler("source_ip_hash", SourceIPHash)
}

// GetSourceAddr is used to get the source addr, which is used by SourceIPHash.
var GetSourceAddr = getSourceAddr

func getSourceAddr(req interface{}) string {
	switch v := req.(type) {
	case *reqresp.Context:
		return v.RemoteAddr

	case *http.Request:
		return v.RemoteAddr

	case interface{ RemoteAddr() net.Addr }:
		return v.RemoteAddr().String()

	default:
		panic(fmt.Errorf("GetSourceAddr: unknown type %T", req))
	}
}

// SourceIPHash returns a new balancer based on the source-ip hash.
//
// The policy name is "source_ip_hash".
func SourceIPHash() Balancer {
	random := newRandom()
	return NewBalancer("source_ip_hash",
		func(c context.Context, r interface{}, sd upstream.ServerDiscovery) error {
			ss := sd.OnServers()
			_len := len(ss)
			if _len == 1 {
				return ss[0].Serve(c, r)
			}

			var value uint64
			host, _ := nets.SplitHostPort(GetSourceAddr(r))
			switch ip := net.ParseIP(host); len(ip) {
			case net.IPv4len:
				value = uint64(binary.BigEndian.Uint32(ip))
			case net.IPv6len:
				value = binary.BigEndian.Uint64(ip[8:16])
			default:
				value = uint64(random(_len))
			}

			return ss[value%uint64(_len)].Serve(c, r)
		})
}
