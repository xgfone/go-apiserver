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
	"net/netip"

	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/nets"
	"github.com/xgfone/go-apiserver/upstream"
)

func init() {
	registerBuiltinBuidler("source_ip_hash", SourceIPHash)
}

// GetSourceAddr is used to get the source addr, which is used by SourceIPHash.
//
// For the default implementation, supports the types of req:
//
//	*http.Request
//	*reqresp.Context
//	interface{ RemoteAddr() string }
//	interface{ RemoteAddr() net.IP }
//	interface{ RemoteAddr() net.Addr }
//	interface{ RemoteAddr() netip.Addr }
var GetSourceAddr func(req interface{}) (netip.Addr, error) = getSourceAddr

func getSourceAddr(req interface{}) (addr netip.Addr, err error) {
	switch v := req.(type) {
	case *reqresp.Context:
		return netip.ParseAddr(v.RemoteAddr)

	case *http.Request:
		return netip.ParseAddr(v.RemoteAddr)

	case interface{ RemoteAddr() string }:
		return netip.ParseAddr(v.RemoteAddr())

	case interface{ RemoteAddr() net.IP }:
		return nets.ToAddr(v.RemoteAddr()), nil

	case interface{ RemoteAddr() net.Addr }:
		return netip.ParseAddr(v.RemoteAddr().String())

	case interface{ RemoteAddr() netip.Addr }:
		return v.RemoteAddr(), nil

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

			ip, err := GetSourceAddr(r)
			if err != nil {
				return err
			}

			var value uint64
			switch ip.BitLen() {
			case 32:
				b4 := ip.As4()
				value = uint64(binary.BigEndian.Uint32(b4[:]))

			case 128:
				b16 := ip.As16()
				value = binary.BigEndian.Uint64(b16[8:16])

			default:
				value = uint64(random(_len))
			}

			return ss[value%uint64(_len)].Serve(c, r)
		})
}
