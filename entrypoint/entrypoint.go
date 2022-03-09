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
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/tcp"
	"github.com/xgfone/go-apiserver/tls/tlscert"
)

// TLSConfig is used to configure the TLS config.
type TLSConfig interface {
	SetTLSConfig(*tls.Config)
	SetTLSForce(forceTLS bool)
}

// Server represents an entrypoint server.
type Server interface {
	TLSConfig
	Protocal() string
	OnShutdown(...func())
	Shutdown(context.Context)
	Start()
}

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
	//
	// For the tcp entrypoint, it must be the type of tcp.Handler.
	// For the http entrypoint, it may be nil or the type of http.Handler.
	// If nil, it is router.NewRouter(ruler.NewRouteManager()) by default.
	Handler interface{}

	Server
}

// NewEntryPoint returns a new entrypoint.
func NewEntryPoint(name, addr string, handler interface{}) *EntryPoint {
	return &EntryPoint{Name: name, Addr: addr, Handler: handler}
}

// Init initializes the entrypoint server.
func (ep *EntryPoint) Init() (err error) {
	if ep.Server != nil {
		return
	}

	addr := ep.Addr
	var protocol string
	if index := strings.Index(addr, "://"); index > -1 {
		protocol = addr[:index]
		addr = addr[index+3:]
	}

	switch protocol {
	case "", "http":
		ln, err := tcp.Listen(addr)
		if err != nil {
			return err
		}

		var httpHandler http.Handler
		switch handler := ep.Handler.(type) {
		case nil:
		case http.Handler:
			httpHandler = handler
		default:
			panic(fmt.Errorf("unknown http handler type '%T'", ep.Handler))
		}

		httpServer := NewHTTPServer(ln, httpHandler)
		ep.Server = httpServer

	case "tcp":
		ln, err := tcp.Listen(addr)
		if err != nil {
			return err
		}

		var tcpHandler tcp.Handler
		switch handler := ep.Handler.(type) {
		case nil:
		case tcp.Handler:
			tcpHandler = handler
		default:
			panic(fmt.Errorf("unknown http handler type '%T'", ep.Handler))
		}

		tcpServer := NewTCPServer(ln, tcpHandler)
		ep.Server = tcpServer

	// case "udp":
	default:
		return fmt.Errorf("unknown entrypoint protocol '%s'", protocol)
	}

	return
}

// Stop is equal to ep.Shutdown(context.Background()).
func (ep *EntryPoint) Stop() { ep.Shutdown(context.Background()) }

// Start starts the entrypoint.
func (ep *EntryPoint) Start() {
	log.Info(fmt.Sprintf("start the %s server", ep.Protocal()),
		"name", ep.Name, "listenaddr", ep.Addr)

	ep.Server.Start()

	log.Info(fmt.Sprintf("stop the %s server", ep.Protocal()),
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
