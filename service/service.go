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

package service

// Service represents a service interface.
type Service interface {
	// Activate is used to activate the service to work in the background.
	Activate()

	// Deactivate is used to deactivate the service to stop the work.
	Deactivate()
}

// Services represents a group services.
type Services []Service

// Append appends the new services into the original group
// and returns the new services.
func (ss Services) Append(services ...Service) Services {
	return append(ss, services...)
}

// Clone clones itself and appends the new services into the cloned services.
func (ss Services) Clone(services ...Service) Services {
	newss := make(Services, len(services)+len(ss))
	copy(newss[len(ss):], services)
	copy(newss, ss)
	return newss
}

// Activate activate all the services in the group.
func (ss Services) Activate() {
	for _, s := range ss {
		s.Activate()
	}
}

// Deactivate deactivates all the service in the group.
func (ss Services) Deactivate() {
	for _, s := range ss {
		s.Deactivate()
	}
}
