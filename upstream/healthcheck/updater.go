// Copyright 2023 xgfone
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

package healthcheck

import (
	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/upstream"
)

// Updater is used to update the server status.
type Updater interface {
	UpsertServer(upstream.Server)
	RemoveServer(serverID string)
	SetServerOnline(serverID string, online bool)
}

// NewLogUpdater returns a proxy updater, which logs the information with the level.
func NewLogUpdater(updater Updater, level log.Level) Updater {
	return logUpdater{updater: updater, level: level}
}

type logUpdater struct {
	updater Updater
	level   log.Level
}

func (u logUpdater) Unwrap() Updater { return u.updater }

func (u logUpdater) UpsertServer(server upstream.Server) {
	log.Log(1, u.level, "upsert upstream server", "serverid", server.ID())
	u.updater.UpsertServer(server)
}

func (u logUpdater) RemoveServer(serverID string) {
	log.Log(1, u.level, "remove upstream server", "serverid", serverID)
	u.updater.RemoveServer(serverID)
}

func (u logUpdater) SetServerOnline(serverID string, online bool) {
	log.Log(1, u.level, "set upstream server online", "serverid", serverID, "online", online)
	u.updater.SetServerOnline(serverID, online)
}
