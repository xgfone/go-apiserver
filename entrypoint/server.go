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

package entrypoint

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-atexit"
)

// TLSConfig is used to configure the TLS config.
type TLSConfig interface {
	SetTLSConfig(*tls.Config)
	SetTLSForce(forceTLS bool)
}

// Server represents an entrypoint server.
type Server interface {
	TLSConfig
	Protocol() string
	OnShutdown(...func())
	Shutdown(context.Context)
	Start()
}

// ServerBuilder is used to build the entrypoint server.
type ServerBuilder func(addr string, handler interface{}) (Server, error)

var builders = make(map[string]ServerBuilder, 4)

// RegisterServerBuilder registers a new builder for the protocol server.
//
// It only registers the "http" protocol server builder by default.
func RegisterServerBuilder(protocol string, builder ServerBuilder) (err error) {
	if protocol == "" {
		panic("the server builder protocol is emtpy")
	} else if builder == nil {
		panic("the server builder is nil")
	}

	if _, ok := builders[protocol]; ok {
		err = fmt.Errorf("the server builder protocol '%s' has existed", protocol)
	} else {
		builders[protocol] = builder
	}

	return
}

// UnregisterServerBuilder unregisters the server builder by the protocol.
func UnregisterServerBuilder(protocol string) { delete(builders, protocol) }

// GetServerBuilder returns the server builder by the protocol.
//
// If the server builder does not exist, return nil.
func GetServerBuilder(protocol string) ServerBuilder { return builders[protocol] }

// GetServerBuilderProtocols returns the protocols of all the server builders.
func GetServerBuilderProtocols() (protocols []string) {
	protocols = make([]string, 0, len(builders))
	for protocol := range builders {
		protocols = append(protocols, protocol)
	}
	return
}

// BuildServer uses the specific protocol server builder to build a server
// with the address and the handler.
func BuildServer(protocol, addr string, handler interface{}) (Server, error) {
	if builder := GetServerBuilder(protocol); builder != nil {
		return builder(addr, handler)
	}
	return nil, fmt.Errorf("no the server builder protocol '%s'", protocol)
}

// StartTLS is used to rapidly start the entrypoint server with TLS.
func StartTLS(name, addr string, handler interface{}, tlsconfig *tls.Config, forceTLS bool) {
	ep, err := NewEntryPoint(name, addr, handler)
	if err != nil {
		log.Fatal("fail to start the server", "name", name, "addr", addr, "err", err)
	}

	atexit.OnExit(ep.Stop)
	ep.OnShutdown(atexit.Execute)
	ep.SetTLSConfig(tlsconfig)
	ep.SetTLSForce(forceTLS)
	ep.Start()
}

// Start is used to rapidly start the entrypoint server.
func Start(addr string, handler interface{}) {
	StartTLS("", addr, handler, nil, false)
}
