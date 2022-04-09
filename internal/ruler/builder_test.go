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

package ruler

import (
	"net/http"
	"testing"
)

func testBuilder(t *testing.T, rule, expect string, req *http.Request, result bool) {
	builder := NewBuilder()
	if matcher, err := builder.Parse(rule); err != nil {
		t.Error(err)
	} else if matcher.String() != expect {
		t.Errorf("expect '%s', but got '%s'", expect, matcher.String())
	} else if ok := matcher.Match(nil, req); ok != result {
		t.Errorf("the rule '%s' does not match the request", rule)
	}
}

func TestMatcherRuleBuilder(t *testing.T) {
	var rule, expect string
	req, _ := http.NewRequest("GET", "http://www.example.com/path", nil)
	req.RemoteAddr = "1.2.3.4"

	rule = "Method(`GET`, `POST`) && Path(`/path`)"
	expect = "And(Or(Method(GET), Method(POST)), Path(/path))"
	testBuilder(t, rule, expect, req, true)

	rule = "Host(`www.example.com`) && Method(`GET`) && Path(`/path`)"
	expect = "And(Host(www.example.com), Method(GET), Path(/path))"
	testBuilder(t, rule, expect, req, true)

	rule = "Method(`GET`, `POST`) || Path(`/path`)"
	expect = "Or(Method(GET), Method(POST), Path(/path))"
	testBuilder(t, rule, expect, req, true)

	rule = "Host(`www.example.com`) || Method(`GET`) || Path(`/path`)"
	expect = "Or(Host(www.example.com), Method(GET), Path(/path))"
	testBuilder(t, rule, expect, req, true)

	rule = "Host(`www.example.com`) && Method(`GET`) || Path(`/path`)"
	expect = "Or(And(Host(www.example.com), Method(GET)), Path(/path))"
	testBuilder(t, rule, expect, req, true)

	rule = "Host(`www.example.com`) || Method(`GET`) && Path(`/path`)"
	expect = "Or(And(Method(GET), Path(/path)), Host(www.example.com))"
	testBuilder(t, rule, expect, req, true)

	rule = "Host(`www.example.com`) && (Method(`GET`) || Path(`/path`))"
	expect = "And(Or(Method(GET), Path(/path)), Host(www.example.com))"
	testBuilder(t, rule, expect, req, true)

	rule = "(Host(`www.example.com`) && ClientIP(`1.2.3.4`)) || (Method(`GET`) && Path(`/path`))"
	expect = "Or(And(Host(www.example.com), ClientIP(1.2.3.4)), And(Method(GET), Path(/path)))"
	testBuilder(t, rule, expect, req, true)

	rule = "(Host(`www.example.com`) || ClientIP(`1.2.3.4`)) && (Method(`GET`) || Path(`/path`))"
	expect = "And(Or(Host(www.example.com), ClientIP(1.2.3.4)), Or(Method(GET), Path(/path)))"
	testBuilder(t, rule, expect, req, true)

	rule = "!Method(`GET`)"
	expect = "Not(Method(GET))"
	testBuilder(t, rule, expect, req, false)

	rule = "!Method(`GET`) && !Path(`/path`)"
	expect = "And(Not(Method(GET)), Not(Path(/path)))"
	testBuilder(t, rule, expect, req, false)

	rule = "!Method(`GET`) || !Path(`/path`)"
	expect = "Or(Not(Method(GET)), Not(Path(/path)))"
	testBuilder(t, rule, expect, req, false)

	rule = "!(Method(`GET`) && Path(`/path`))"
	expect = "Or(Not(Method(GET)), Not(Path(/path)))"
	testBuilder(t, rule, expect, req, false)

	rule = "!(Method(`GET`) || Path(`/path`))"
	expect = "And(Not(Method(GET)), Not(Path(/path)))"
	testBuilder(t, rule, expect, req, false)

	rule = "!Method(`GET`) || Path(`/path`)"
	expect = "Or(Not(Method(GET)), Path(/path))"
	testBuilder(t, rule, expect, req, true)
}

func checkHosts(t *testing.T, err error, results, expects []string) {
	if err != nil {
		t.Error(err)
		return
	}

	if len(results) != len(expects) {
		t.Errorf("expect %d hosts, but got %d", len(expects), len(results))
		return
	}

	for i, host := range results {
		if host != expects[i] {
			t.Errorf("expect host '%s', but got '%s'", expects[i], host)
		}
	}
}

func TestParseHostSNI(t *testing.T) {
	hosts, err := ParseHostSNI("HostSNI(`www.example1.com`, `*.example2.com`)")
	checkHosts(t, err, hosts, []string{"www.example1.com", "*.example2.com"})
}

func TestParseDomains(t *testing.T) {
	hosts, err := ParseDomains("Host(`www.example1.com`, `*.example2.com`)")
	checkHosts(t, err, hosts, []string{"www.example1.com", "*.example2.com"})
}
