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
	"context"
	"fmt"
	"strings"

	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/tls/tlscert"
)

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

// NewEntryPoint returns a new entrypoint.
func NewEntryPoint(name, addr string, handler interface{}) *EntryPoint {
	return &EntryPoint{Name: name, Addr: addr, Handler: handler}
}

// Init initializes the entrypoint server, which extracts the protocol
// from the address and builds the server by the protocol server builder.
func (ep *EntryPoint) Init() (err error) {
	if ep.Server != nil {
		return
	}

	addr := ep.Addr
	protocol := "http"
	if index := strings.Index(addr, "://"); index > -1 {
		protocol = addr[:index]
		addr = addr[index+3:]
	}

	ep.Server, err = BuildServer(protocol, addr, ep.Handler)
	return
}

// Stop is equal to ep.Shutdown(context.Background()).
func (ep *EntryPoint) Stop() { ep.Shutdown(context.Background()) }

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
