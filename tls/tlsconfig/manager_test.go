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

package tlsconfig

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/xgfone/go-apiserver/internal/test"
)

type testUpdater struct {
	buf *bytes.Buffer
}

func (u testUpdater) AddTLSConfig(name string, config *tls.Config) {
	fmt.Fprintf(u.buf, "add the tls config named '%s'\n", name)
}

func (u testUpdater) DelTLSConfig(name string) {
	fmt.Fprintf(u.buf, "delete the tls config named '%s'\n", name)
}

func TestManager(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	updaters := NewNamedUpdaters()
	updaters.AddUpdater("updater1", testUpdater{buf})

	m := NewManager()
	m.SetUpdater(updaters)
	m.AddTLSConfig("tlsconfig1", new(tls.Config))
	m.AddTLSConfig("tlsconfig2", new(tls.Config))
	m.DelTLSConfig("tlsconfig1")

	results := strings.Split(buf.String(), "\n")
	expects := []string{
		"add the tls config named 'tlsconfig1'",
		"add the tls config named 'tlsconfig2'",
		"delete the tls config named 'tlsconfig1'",
		"",
	}

	sort.Strings(results)
	sort.Strings(expects)
	test.CheckStrings(t, "TLSConfigManager", results, expects)
}
