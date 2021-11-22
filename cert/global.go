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

package cert

import "context"

// DefaultCertManagerGroup is the default certificate manager group.
var DefaultCertManagerGroup = NewCertManagerGroup()

// AddCertManager is equal to DefaultCertManagerGroup.AddCertManager(cm).
func AddCertManager(cm *CertManager) error {
	return DefaultCertManagerGroup.AddCertManager(cm)
}

// DelCertManager is equal to DefaultCertManagerGroup.DelCertManager(name).
func DelCertManager(name string) {
	DefaultCertManagerGroup.DelCertManager(name)
}

// GetCertManager is equal to DefaultCertManagerGroup.GetCertManager(name).
func GetCertManager(name string) *CertManager {
	return DefaultCertManagerGroup.GetCertManager(name)
}

// GetCertManagers is equal to DefaultCertManagerGroup.GetCertManagers().
func GetCertManagers() []*CertManager {
	return DefaultCertManagerGroup.GetCertManagers()
}

/// ----------------------------------------------------------------------- ///

// DefaultCertManager is the default certificate manager.
var DefaultCertManager = NewCertManager("default")

// GetCertificates is equal to DefaultCertManager.GetCertificates().
func GetCertificates() map[string]Certificate {
	return DefaultCertManager.GetCertificates()
}

// GetCertificate is equal to DefaultCertManager.GetCertificate(name).
func GetCertificate(name string) (cert Certificate, ok bool) {
	return DefaultCertManager.GetCertificate(name)
}

// AddCertificate is equal to DefaultCertManager.AddCertificate(name, cert).
func AddCertificate(name string, cert Certificate) {
	DefaultCertManager.AddCertificate(name, cert)
}

// DelCertificate is equal to DefaultCertManager.DelCertificate(name).
func DelCertificate(name string) { DefaultCertManager.DelCertificate(name) }

/// ----------------------------------------------------------------------- ///

// DefaultProviderManager is the default certificate provider manager.
var DefaultProviderManager = NewProviderManager(DefaultCertManager)

// StartProviderManager is equal to DefaultProviderManager.Start(ctx).
func StartProviderManager(ctx context.Context) {
	DefaultProviderManager.Start(ctx)
}

// StopProviderManager is equal to DefaultProviderManager.Stop().
func StopProviderManager() { DefaultProviderManager.Stop() }

// GetProviders is equal to DefaultProviderManager.GetProviders().
func GetProviders() []Provider {
	return DefaultProviderManager.GetProviders()
}

// GetProvider is equal to DefaultProviderManager.GetProvider(name).
func GetProvider(name string) Provider {
	return DefaultProviderManager.GetProvider(name)
}

// AddProvider is equal to DefaultProviderManager.AddProvider(provider).
func AddProvider(provider Provider) (err error) {
	return DefaultProviderManager.AddProvider(provider)
}

// DelProvider is equal to DefaultProviderManager.DelProvider(name).
func DelProvider(name string) { DefaultProviderManager.DelProvider(name) }
