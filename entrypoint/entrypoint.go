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

// Package entrypoint implements the http entrypoint and the manager.
package entrypoint

import (
	"fmt"
	"strings"

	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/tls/tlscert"
)

// EntryPointHook is used by NewEntryPoint to intercept the built entrypoint.
//
// Default: nil
var EntryPointHook func(*EntryPoint)

// EntryPoint represents an entrypoint of the services.
type EntryPoint struct {
	// The unique name of the entrypoint.
	Name string

	// The address that the entrypoint listens on.
	//
	// Format: [(http|tcp|udp)://][host]:port
	//
	// If missing the protocol, it is "http" by default.
	Addr string

	// Handler is the handler of the entrypoint server.
	Handler interface{}

	Server
}

// NewEntryPoint news an entrypoint, which extracts the protocol from addr
// and builds the server by the protocol server builder.
func NewEntryPoint(name, addr string, handler interface{}) (*EntryPoint, error) {
	protocol := "http"
	if index := strings.Index(addr, "://"); index > -1 {
		protocol = addr[:index]
		addr = addr[index+3:]
	}

	server, err := BuildServer(protocol, addr, handler)
	if err != nil {
		return nil, err
	}

	ep := &EntryPoint{Name: name, Addr: addr, Handler: handler, Server: server}
	if EntryPointHook != nil {
		EntryPointHook(ep)
	}
	return ep, nil
}

// Start starts the entrypoint.
func (ep *EntryPoint) Start() {
	log.Info(fmt.Sprintf("start the %s server", ep.Protocol()),
		"name", ep.Name, "listenaddr", ep.Addr)

	ep.Server.Start()

	log.Info(fmt.Sprintf("stop the %s server", ep.Protocol()),
		"name", ep.Name, "listenaddr", ep.Addr)
}

// AddCertificate implements the interface tlscert.CertUpdater
// to add the certificate with the name into the server if it supports TLS.
func (ep *EntryPoint) AddCertificate(name string, certificate tlscert.Certificate) {
	if updater, ok := ep.Server.(tlscert.Updater); ok {
		updater.AddCertificate(name, certificate)
	}
}

// DelCertificate implements the interface cert.CertUpdater
// to delete the certificate by the name from the server if it supports TLS.
func (ep *EntryPoint) DelCertificate(name string) {
	if updater, ok := ep.Server.(tlscert.Updater); ok {
		updater.DelCertificate(name)
	}
}
