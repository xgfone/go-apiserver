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

package tls

import (
	"bytes"
	"crypto/tls"
	"strings"
	"testing"

	"github.com/xgfone/go-apiserver/internal/test"
	"github.com/xgfone/go-apiserver/tls/tlscert"
	"github.com/xgfone/go-apiserver/tls/tlsconfig"
)

func TestClientConfig(t *testing.T) {
	cert, _ := tlscert.NewCertificate([]byte(test.Cert), []byte(test.Key))
	buf := bytes.NewBuffer(nil)
	c := NewClientConfig(nil)
	c.OnUpdate(tlsconfig.SetterFunc(func(*tls.Config) { buf.WriteString("update tls.Config\n") }))

	c.AddCertificate("certname", cert)
	c.SetTLSConfig(new(tls.Config))

	results := strings.Split(buf.String(), "\n")
	test.CheckStrings(t, "TestClientConfig", results, []string{
		"update tls.Config",
		"update tls.Config",
		"",
	})
}

func TestServerConfig(t *testing.T) {
	cert, _ := tlscert.NewCertificate([]byte(test.Cert), []byte(test.Key))
	buf := bytes.NewBuffer(nil)
	c := NewServerConfig(nil)
	c.OnUpdate(tlsconfig.SetterFunc(func(*tls.Config) { buf.WriteString("update tls.Config\n") }))

	c.AddCertificate("certname", cert)
	c.SetTLSConfig(new(tls.Config))

	results := strings.Split(buf.String(), "\n")
	test.CheckStrings(t, "TestServerConfig", results, []string{
		"update tls.Config",
		"",
	})
}
