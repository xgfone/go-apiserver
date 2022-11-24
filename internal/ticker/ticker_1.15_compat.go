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

//go:build !go1.15
// +build !go1.15

package ticker

import (
	"time"

	"github.com/xgfone/go-apiserver/helper"
)

// Ticker is the replacer of time.Ticker.
type Ticker struct{ *time.Ticker }

// NewTicker is used to new a ticker.
func NewTicker(d time.Duration) *Ticker { return &Ticker{time.NewTicker(d)} }

// Stop stops the ticker.
func (t *Ticker) Stop() { helper.StopTicker(t.Ticker) }

// Reset is the simple implementation of time.Ticker#Reset.
func (t *Ticker) Reset(d time.Duration) {
	helper.StopTicker(t.Ticker)
	t.Ticker = time.NewTicker(d)
}
