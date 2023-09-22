// Copyright 2022~2023 xgfone
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

import (
	"log/slog"
)

// Log returns a new Service, which will log the activated or deactivated event
// with the service name.
func Log(name string, service Service) Service {
	return logService{service: service, name: name}
}

type logService struct {
	service Service
	name    string
}

func (s logService) Activate() {
	slog.Info("the service is activated", "service", s.name)
	s.service.Activate()
}

func (s logService) Deactivate() {
	slog.Info("the service is deactivated", "service", s.name)
	s.service.Deactivate()
}
