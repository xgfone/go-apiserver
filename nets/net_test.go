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

package nets

import (
	"fmt"
	"net"
	"testing"
)

func TestIsTimeout(t *testing.T) {
	if IsTimeout(fmt.Errorf("error")) {
		t.Error("unexpect timeout")
	}

	if IsTimeout(&net.DNSError{IsTimeout: false}) {
		t.Error("unexpect timeout")
	}

	if !IsTimeout(&net.DNSError{IsTimeout: true}) {
		t.Error("expect timeout")
	}
}

func ExampleNormalizeMac() {
	fmt.Println(NormalizeMac("00:aa:bb:cc:dd:ee"))
	fmt.Println(NormalizeMac("00:AA:BB:CC:DD:EE"))
	fmt.Println(NormalizeMac("00:Aa:Bb:Cc:Dd:Ee"))
	fmt.Println(NormalizeMac("00-aa-bb-cc-dd-ee"))
	fmt.Println(NormalizeMac("00-AA-BB-CC-DD-EE"))
	fmt.Println(NormalizeMac("00-Aa-Bb-Cc-Dd-Ee"))
	fmt.Println(NormalizeMac("00aa.bbcc.ddee"))
	fmt.Println(NormalizeMac("00AA.BBCC.DDEE"))
	fmt.Println(NormalizeMac("00Aa.BbCc.DdEe"))

	// Output:
	// 00:aa:bb:cc:dd:ee
	// 00:aa:bb:cc:dd:ee
	// 00:aa:bb:cc:dd:ee
	// 00:aa:bb:cc:dd:ee
	// 00:aa:bb:cc:dd:ee
	// 00:aa:bb:cc:dd:ee
	// 00:aa:bb:cc:dd:ee
	// 00:aa:bb:cc:dd:ee
	// 00:aa:bb:cc:dd:ee
}

func ExampleSplitHostPort() {
	var host, port string

	host, port = SplitHostPort(":80")
	fmt.Printf("host=%s, port=%s\n", host, port)

	host, port = SplitHostPort("1.2.3.4")
	fmt.Printf("host=%s, port=%s\n", host, port)

	host, port = SplitHostPort("1.2.3.4:")
	fmt.Printf("host=%s, port=%s\n", host, port)

	host, port = SplitHostPort("1.2.3.4:80")
	fmt.Printf("host=%s, port=%s\n", host, port)

	host, port = SplitHostPort("[fe80::215:5dff:fe34:60]")
	fmt.Printf("host=%s, port=%s\n", host, port)

	host, port = SplitHostPort("[fe80::215:5dff:fe34:60]:")
	fmt.Printf("host=%s, port=%s\n", host, port)

	host, port = SplitHostPort("[fe80::215:5dff:fe34:60]:80")
	fmt.Printf("host=%s, port=%s\n", host, port)

	// We don't check the validity of the host, so don't use this format.
	host, port = SplitHostPort("fe80::215:5dff:fe34:8e60")
	fmt.Printf("host=%s, port=%s\n", host, port)

	// We don't check the validity of the host, so don't use this format.
	host, port = SplitHostPort("fe80::215:5dff:fe34:8e60:")
	fmt.Printf("host=%s, port=%s\n", host, port)

	// We don't check the validity of the host, so don't use this format.
	host, port = SplitHostPort("fe80::215:5dff:fe34:8e60:80")
	fmt.Printf("host=%s, port=%s\n", host, port)

	// Output:
	// host=, port=80
	// host=1.2.3.4, port=
	// host=1.2.3.4, port=
	// host=1.2.3.4, port=80
	// host=fe80::215:5dff:fe34:60, port=
	// host=fe80::215:5dff:fe34:60, port=
	// host=fe80::215:5dff:fe34:60, port=80
	// host=fe80::215:5dff:fe34:8e60, port=
	// host=fe80::215:5dff:fe34:8e60, port=
	// host=fe80::215:5dff:fe34:8e60, port=80
}

func ExampleIPChecker() {
	checkip := func(checker IPChecker, ip string) {
		if checker.CheckIP(net.ParseIP(ip)) {
			fmt.Printf("'%v' contains the ip '%s'\n", checker, ip)
		} else {
			fmt.Printf("'%v' does not contain the ip '%s'\n", checker, ip)
		}
	}

	ipv4checker, _ := NewIPChecker("1.2.3.4")
	checkip(ipv4checker, "1.2.3.4")
	checkip(ipv4checker, "5.6.7.8")
	checkip(ipv4checker, "fe80::215:5dff:fe8c:6de7")

	cidrv4checker, _ := NewIPChecker("10.0.0.0/8")
	checkip(cidrv4checker, "10.1.2.3")
	checkip(cidrv4checker, "192.168.1.2")
	checkip(cidrv4checker, "fe80::215:5dff:fe8c:6de7")

	ipv6checker, _ := NewIPChecker("fe80::215:5dff:fe8c:6de7")
	checkip(ipv6checker, "fe80::215:5dff:fe8c:6de7")
	checkip(ipv6checker, "fe80::215:5dff:fe8c:1234")
	checkip(ipv6checker, "1.2.3.4")

	cidrv6checker, _ := NewIPChecker("fe80::/16")
	checkip(cidrv6checker, "fe80::215:5dff:fe8c:6de7")
	checkip(cidrv6checker, "2408::215:5dff:fe8c:6de7")
	checkip(cidrv6checker, "1.2.3.4")

	checkers, _ := NewIPCheckers("10.0.0.0/8", "fe80::/16")
	checkip(checkers, "10.1.2.3")
	checkip(checkers, "192.168.1.2")
	checkip(checkers, "fe80::215:5dff:fe8c:6de7")
	checkip(checkers, "2408::215:5dff:fe8c:6de7")

	// Output:
	// '1.2.3.4/32' contains the ip '1.2.3.4'
	// '1.2.3.4/32' does not contain the ip '5.6.7.8'
	// '1.2.3.4/32' does not contain the ip 'fe80::215:5dff:fe8c:6de7'
	// '10.0.0.0/8' contains the ip '10.1.2.3'
	// '10.0.0.0/8' does not contain the ip '192.168.1.2'
	// '10.0.0.0/8' does not contain the ip 'fe80::215:5dff:fe8c:6de7'
	// 'fe80::215:5dff:fe8c:6de7/128' contains the ip 'fe80::215:5dff:fe8c:6de7'
	// 'fe80::215:5dff:fe8c:6de7/128' does not contain the ip 'fe80::215:5dff:fe8c:1234'
	// 'fe80::215:5dff:fe8c:6de7/128' does not contain the ip '1.2.3.4'
	// 'fe80::/16' contains the ip 'fe80::215:5dff:fe8c:6de7'
	// 'fe80::/16' does not contain the ip '2408::215:5dff:fe8c:6de7'
	// 'fe80::/16' does not contain the ip '1.2.3.4'
	// '10.0.0.0/8,fe80::/16' contains the ip '10.1.2.3'
	// '10.0.0.0/8,fe80::/16' does not contain the ip '192.168.1.2'
	// '10.0.0.0/8,fe80::/16' contains the ip 'fe80::215:5dff:fe8c:6de7'
	// '10.0.0.0/8,fe80::/16' does not contain the ip '2408::215:5dff:fe8c:6de7'
}
