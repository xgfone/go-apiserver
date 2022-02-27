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

package tlscert

import (
	"context"
	"fmt"
	"sync"

	"github.com/xgfone/go-apiserver/log"
)

// CertUpdater is used to update the certificates.
type CertUpdater interface {
	AddCertificate(name string, cert Certificate)
	DelCertificate(name string)
}

type noopCertUpdater struct{}

func (u noopCertUpdater) AddCertificate(name string, cert Certificate) {}
func (u noopCertUpdater) DelCertificate(name string)                   {}

// LogCertUpdater returns a new certificate updater to log the change
// of the certificates.
func LogCertUpdater(u CertUpdater) CertUpdater { return logCertUpdater{u} }

type logCertUpdater struct{ updater CertUpdater }

func (u logCertUpdater) AddCertificate(name string, cert Certificate) {
	log.Info("add the certificate", "name", name)
	if u.updater != nil {
		u.updater.AddCertificate(name, cert)
	}
}

func (u logCertUpdater) DelCertificate(name string) {
	log.Info("delete the certificate", "name", name)
	if u.updater != nil {
		u.updater.DelCertificate(name)
	}
}

// CertUpdaters is a set of CertUpdaters.
type CertUpdaters []CertUpdater

// AddCertificate implements the interface CertUpdater,
// which will forward to each updater member.
func (us CertUpdaters) AddCertificate(name string, cert Certificate) {
	for _, updater := range us {
		updater.AddCertificate(name, cert)
	}
}

// DelCertificate implements the interface CertUpdater,
// which will forward to each updater member.
func (us CertUpdaters) DelCertificate(name string) {
	for _, updater := range us {
		updater.DelCertificate(name)
	}
}

// Provider is used to provide the certificates.
type Provider interface {
	// Name is the name of the provider, which indicates uniquely the provider.
	Name() string

	// OnChanged is used to update the certificate when the certificate
	// has changed.
	//
	// If context is done, the provider should clean the resources and stop.
	OnChanged(context.Context, CertUpdater)
}

type providerWrapper struct {
	context.CancelFunc
	Provider
}

// ProviderManager is used to manage a set of the certificate providers.
type ProviderManager struct {
	updater CertUpdater

	lock      sync.RWMutex
	cancel    context.CancelFunc
	context   context.Context
	providers map[string]*providerWrapper
}

// NewProviderManager returns a new provider manager with the updater.
func NewProviderManager(updater CertUpdater) *ProviderManager {
	return &ProviderManager{
		updater:   updater,
		providers: make(map[string]*providerWrapper, 4),
	}
}

// GetProviders returns all the certificate providers.
func (pm *ProviderManager) GetProviders() []Provider {
	pm.lock.RLock()
	providers := make([]Provider, 0, len(pm.providers))
	for _, provider := range pm.providers {
		providers = append(providers, provider)
	}
	pm.lock.RUnlock()
	return providers
}

// GetProvider returns the certificate provider by the name.
func (pm *ProviderManager) GetProvider(name string) Provider {
	pm.lock.RLock()
	provider := pm.providers[name]
	pm.lock.RUnlock()
	return provider
}

// AddProvider adds the provider into the manager.
func (pm *ProviderManager) AddProvider(provider Provider) (err error) {
	var pw *providerWrapper
	name := provider.Name()

	pm.lock.Lock()
	defer pm.lock.Unlock()

	if pm.providers == nil {
		pm.providers = make(map[string]*providerWrapper)
	}

	if _, ok := pm.providers[name]; ok {
		err = fmt.Errorf("the certificate provider named '%s' has existed", name)
	} else {
		pw = &providerWrapper{Provider: provider}
		pm.providers[name] = pw
	}

	// The provider manager has been started.
	if pm.context != nil {
		pm.startProvider(pw)
	}

	return
}

// DelProvider stops and deletes the certificate provider by the name.
//
// If the provider does not exist, do nothing.
func (pm *ProviderManager) DelProvider(name string) {
	pm.lock.Lock()
	defer pm.lock.Unlock()

	if len(pm.providers) != 0 {
		if provider, ok := pm.providers[name]; ok {
			delete(pm.providers, name)
			provider.CancelFunc()
		}
	}
}

func (pm *ProviderManager) startProvider(pw *providerWrapper) {
	var ctx context.Context
	ctx, pw.CancelFunc = context.WithCancel(pm.context)
	go pw.OnChanged(ctx, pm.updater)
}

// Start starts all the providers.
func (pm *ProviderManager) Start(ctx context.Context) {
	pm.lock.Lock()
	defer pm.lock.Unlock()

	if pm.context != nil { // Has been started
		return
	}

	pm.context, pm.cancel = context.WithCancel(ctx)
	for _, provider := range pm.providers {
		pm.startProvider(provider)
	}
}

// Stop stops all the providers.
func (pm *ProviderManager) Stop() {
	pm.lock.Lock()
	defer pm.lock.Unlock()

	if pm.context == nil { // Has not been started
		return
	}

	pm.cancel()
	pm.providers = nil
	pm.context, pm.cancel = nil, nil
}
