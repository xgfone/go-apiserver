// Copyright 2021~2022 xgfone
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

// Package nets provides some convenient net functions.
package nets

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"
)

type timeoutError interface {
	Timeout() bool // Is the error a timeout?
	error
}

// IsTimeout reports whether the error is timeout.
func IsTimeout(err error) bool {
	var timeoutErr timeoutError
	return errors.As(err, &timeoutErr) && timeoutErr.Timeout()
}

// IsClosed reports whether the error is closed.
func IsClosed(err error) bool { return errors.Is(err, net.ErrClosed) }

// NormalizeMac normalizes the mac, which is the convenient function
// of net.ParseMAC, but only supports the 48-bit format and outputs
// the string like "xx:xx:xx:xx:xx:xx".
//
// Return "" if the mac is an invalid mac.
func NormalizeMac(mac string) string {
	if ha, err := net.ParseMAC(mac); err == nil || len(ha) == 6 {
		return ha.String()
	}
	return ""
}

// ToIP converts any value to net.IP, but returns nil if failing.
func ToIP(v interface{}) net.IP {
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

// IPIsOnInterface reports whether the ip is on the given network interface
// named ifaceName.
//
// If ip is empty or invalid, return false.
// If ifaceName is empty, it checks all the network interfaces.
func IPIsOnInterface(ip, ifaceName string) (on bool, err error) {
	netip := net.ParseIP(strings.TrimSpace(ip))
	if netip == nil {
		return false, nil
	}

	var addrs []net.Addr
	var iface *net.Interface
	if ifaceName == "" {
		addrs, err = net.InterfaceAddrs()
	} else if iface, err = net.InterfaceByName(ifaceName); err == nil {
		addrs, err = iface.Addrs()
	}

	if err != nil {
		return
	}

	ip = netip.String()
	for _, addr := range addrs {
		if strings.Split(addr.String(), "/")[0] == ip {
			return true, nil
		}
	}

	return false, nil
}

// SplitHostPort separates host and port. If the port is not valid, it returns
// the entire input as host, and it doesn't check the validity of the host.
// Unlike net.SplitHostPort, but per RFC 3986, it requires ports to be numeric.
func SplitHostPort(hostport string) (host, port string) {
	host = hostport

	colon := strings.LastIndexByte(host, ':')
	if colon != -1 && validOptionalPort(host[colon:]) {
		host, port = host[:colon], host[colon+1:]
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}

	return
}

// validOptionalPort reports whether port is either an empty string
// or matches /^:\d*$/
func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

// IPChecker is used to check whether the ip is legal or allowed.
type IPChecker interface {
	CheckIP(ip net.IP) (ok bool)
}

// IPCheckerFunc is the ip checker function.
type IPCheckerFunc func(net.IP) bool

// CheckIP implements the interface IPChecker.
func (c IPCheckerFunc) CheckIP(ip net.IP) bool { return c(ip) }

// IPCheckers is a set of IPChecker.
type IPCheckers []IPChecker

// NewIPCheckers parses a group of the ip or cidr strings to IPCheckers.
func NewIPCheckers(ipOrCidrs ...string) (IPCheckers, error) {
	checkers := make(IPCheckers, len(ipOrCidrs))
	for i, ip := range ipOrCidrs {
		c, err := NewIPChecker(ip)
		if err != nil {
			return nil, err
		}
		checkers[i] = c
	}
	return checkers, nil
}

func (cs IPCheckers) String() string {
	_len := len(cs)
	if _len == 0 {
		return ""
	}

	var buf strings.Builder
	buf.Grow(128)
	for i := 0; i < _len; i++ {
		if i > 0 {
			buf.WriteByte(',')
		}

		if s, ok := cs[i].(fmt.Stringer); ok {
			buf.WriteString(s.String())
		} else {
			fmt.Fprint(&buf, cs[i])
		}
	}
	return buf.String()
}

// CheckIP implements the interface IPChecker, which returns true
// if any ip checker return true.
func (cs IPCheckers) CheckIP(ip net.IP) bool {
	if len(cs) == 0 {
		return true
	}

	if len(ip) == 0 {
		return false
	}

	for i, _len := 0, len(cs); i < _len; i++ {
		if cs[i].CheckIP(ip) {
			return true
		}
	}
	return false
}

// NewIPChecker returns a new IPChecker based on an IP or CIDR.
func NewIPChecker(ipOrCidr string) (IPChecker, error) {
	cidr := ipOrCidr

	var isip bool
	if isip = strings.IndexByte(cidr, '/') < 0; isip {
		if strings.IndexByte(cidr, '.') == -1 { // For IPv6
			cidr += "/128"
		} else { // For IPv4
			cidr += "/32"
		}
	}

	ipChecker, err := newIPChecker(cidr)
	if err == nil {
		return ipChecker, nil
	}

	if isip {
		err = fmt.Errorf("invalid ip address '%s': %w", ipOrCidr, err)
	} else {
		err = fmt.Errorf("invalid cidr network address '%s': %w", ipOrCidr, err)
	}

	return nil, err
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
