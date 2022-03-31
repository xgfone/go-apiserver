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

package leaderelection

import (
	"context"
	"fmt"
)

func ExampleLeaderElector() {
	vip := "127.0.0.1"
	config := ElectionConfig{
		Identity: vip,
		Lock:     NewVipResourceLock(vip),
		Callbacks: LeaderCallbacks{
			OnStartedLeading: func(context.Context) {
				fmt.Println("start leading")
			},

			OnStoppedLeading: func() {
				fmt.Println("stop leading")
			},

			OnNewLeader: func(identity string) {
				fmt.Printf("new leader '%s'\n", identity)
			},
		},
		ReleaseOnCancel: true,
	}

	elector, err := NewLeaderElector(config)
	if err != nil {
		fmt.Println(err)
		return
	}

	elector.Run(context.Background())
}
