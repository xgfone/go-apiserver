// Copyright 2022 xgfone
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

//go:build !go1.18
// +build !go1.18

package nets

import (
	"fmt"
	"net"
)

func toIP(v interface{}) net.IP {
	switch ip := v.(type) {
	case string:
		return net.ParseIP(ip)

	case net.IP:
		return ip

	case net.IPAddr:
		return ip.IP

	case net.TCPAddr:
		return ip.IP

	case *net.TCPAddr:
		return ip.IP

	case net.UDPAddr:
		return ip.IP

	case *net.UDPAddr:
		return ip.IP

	case net.Addr:
		s, _ := SplitHostPort(ip.String())
		return net.ParseIP(s)

	case fmt.Stringer:
		return net.ParseIP(ip.String())

	default:
		return nil
	}
}

type ipChecker struct{ *net.IPNet }

func newIPChecker(cidr string) (c ipChecker, err error) {
	_, c.IPNet, err = net.ParseCIDR(cidr)
	return
}

func (c ipChecker) String() string { return c.IPNet.String() }

func (c ipChecker) CheckIP(ip net.IP) (ok bool) {
	if len(ip) > 0 {
		ok = c.IPNet.Contains(ip)
	}
	return
}
