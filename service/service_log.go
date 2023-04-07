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

import "github.com/xgfone/go-apiserver/log"

// LogService returns a new log Service, which will log the activated
// or deactivated event.
func LogService(logLevel log.Level, serviceName string, service Service) Service {
	return logService{service: service, level: logLevel, sname: serviceName}
}

type logService struct {
	service Service
	sname   string
	level   log.Level
}

func (s logService) Activate() {
	log.Log(0, s.level, "the service is activated", "service", s.sname)
	s.service.Activate()
}

func (s logService) Deactivate() {
	log.Log(0, s.level, "the service is deactivated", "service", s.sname)
	s.service.Deactivate()
}
