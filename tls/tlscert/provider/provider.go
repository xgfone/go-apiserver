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

// Package provider provides some tls certificate provider and the provider manager.
package provider

import (
	"context"
	"fmt"
	"sync"

	"github.com/xgfone/go-apiserver/tls/tlscert"
)

// DefaultManager is the default certificate provider manager.
var DefaultManager *Manager

// Provider is used to provide the certificates.
type Provider interface {
	// OnChangedCertificate is used to update the certificate
	// when the certificate has changed.
	//
	// If context is done, the provider should clean the resources and stop.
	OnChangedCertificate(context.Context, tlscert.Updater)
}

type providerWrapper struct {
	context.CancelFunc
	Provider
}

// Manager is used to manage a set of the certificate providers.
type Manager struct {
	updater tlscert.Updater

	lock      sync.RWMutex
	cancel    context.CancelFunc
	context   context.Context
	providers map[string]*providerWrapper
}

// NewManager returns a new provider manager with the updater.
func NewManager(updater tlscert.Updater) *Manager {
	return &Manager{
		updater:   updater,
		providers: make(map[string]*providerWrapper, 4),
	}
}

// GetProviders returns all the certificate providers.
func (m *Manager) GetProviders() map[string]Provider {
	m.lock.RLock()
	providers := make(map[string]Provider, len(m.providers))
	for name, provider := range m.providers {
		providers[name] = provider
	}
	m.lock.RUnlock()
	return providers
}

// GetProvider returns the certificate provider by the name.
//
// Return nil if the provider does not exist.
func (m *Manager) GetProvider(name string) Provider {
	m.lock.RLock()
	provider := m.providers[name]
	m.lock.RUnlock()
	return provider
}

// AddProvider adds the provider named name into the manager.
func (m *Manager) AddProvider(name string, provider Provider) (err error) {
	if name == "" {
		panic("the certificate provider name is empty")
	} else if provider == nil {
		panic("the certificate provider is nil")
	}

	m.lock.Lock()
	if _, ok := m.providers[name]; ok {
		err = fmt.Errorf("the certificate provider named '%s' has existed", name)
	} else {
		pw := &providerWrapper{Provider: provider}
		m.providers[name] = pw

		// The provider manager has been started.
		if m.context != nil {
			m.startProvider(pw)
		}
	}
	m.lock.Unlock()
	return
}

// DelProvider stops and deletes the certificate provider by the name.
//
// If the provider does not exist, do nothing.
func (m *Manager) DelProvider(name string) {
	m.lock.Lock()
	if provider, ok := m.providers[name]; ok {
		delete(m.providers, name)
		provider.CancelFunc()
	}
	m.lock.Unlock()
}

func (m *Manager) startProvider(pw *providerWrapper) {
	var ctx context.Context
	ctx, pw.CancelFunc = context.WithCancel(m.context)
	go pw.OnChangedCertificate(ctx, m.updater)
}

// Start starts all the providers.
func (m *Manager) Start(ctx context.Context) {
	m.lock.Lock()
	if m.context == nil {
		m.context, m.cancel = context.WithCancel(ctx)
		for _, provider := range m.providers {
			m.startProvider(provider)
		}
	}
	m.lock.Unlock()
}

// Stop stops all the providers.
func (m *Manager) Stop() {
	m.lock.Lock()
	if m.context != nil {
		m.cancel()
		m.context, m.cancel = nil, nil
	}
	m.lock.Unlock()
}
