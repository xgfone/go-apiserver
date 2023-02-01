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

//go:build go1.18
// +build go1.18

package nets

import (
	"fmt"
	"net"
	"net/netip"
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

	case netip.Addr:
		return net.IP(ip.AsSlice())

	case net.Addr:
		s, _ := SplitHostPort(ip.String())
		return net.ParseIP(s)

	case fmt.Stringer:
		return net.ParseIP(ip.String())

	default:
		return nil
	}
}

func ip2addr(ip net.IP) (addr netip.Addr) {
	switch len(ip) {
	case net.IPv4len:
		var b [4]byte
		copy(b[:], ip)
		addr = netip.AddrFrom4(b)

	case net.IPv6len:
		var b [16]byte
		copy(b[:], ip)
		addr = netip.AddrFrom16(b)
	}
	return
}

// ToAddr converts any value to netip.Addr, but returns ZERO if failing.
func ToAddr(v interface{}) netip.Addr {
	var addr netip.Addr
	switch ip := v.(type) {
	case string:
		addr, _ = netip.ParseAddr(ip)

	case net.IP:
		addr = ip2addr(ip)

	case net.IPAddr:
		addr = ip2addr(ip.IP)

	case net.TCPAddr:
		addr = ip2addr(ip.IP)

	case *net.TCPAddr:
		addr = ip2addr(ip.IP)

	case net.UDPAddr:
		addr = ip2addr(ip.IP)

	case *net.UDPAddr:
		addr = ip2addr(ip.IP)

	case netip.Addr:
		addr = ip

	case net.Addr:
		s, _ := SplitHostPort(ip.String())
		addr, _ = netip.ParseAddr(s)

	case fmt.Stringer:
		addr, _ = netip.ParseAddr(ip.String())

	}

	return addr
}

type ipChecker struct{ netip.Prefix }

func newIPChecker(cidr string) (c ipChecker, err error) {
	c.Prefix, err = netip.ParsePrefix(cidr)
	return
}

func (c ipChecker) String() string { return c.Prefix.String() }

func (c ipChecker) CheckIP(ip net.IP) bool {
	if len(ip) == 0 {
		return false
	}

	if c.Prefix.Addr().BitLen() == 32 {
		ip = ip.To4()
	} else {
		ip = ip.To16()
	}

	addr, ok := netip.AddrFromSlice(ip)
	return ok && c.Prefix.Contains(addr)
}
