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

package helper

import "time"

// Some common durations.
const (
	Day  = time.Hour * 24
	Week = Day * 7
)

// StopTimer stops the timer.
func StopTimer(timer *time.Timer) {
	if timer != nil {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}
}

// StopTicker stops the time ticker.
func StopTicker(ticker *time.Ticker) {
	if ticker != nil {
		ticker.Stop()
		for {
			select {
			case <-ticker.C:
			default:
				return
			}
		}
	}
}

// TimeAdd adds a duration to t and returns a new time.Time.
func TimeAdd(t time.Time, years, months, days, hours, minutes, seconds int) time.Time {
	if years > 0 || months > 0 || days > 0 {
		t = t.AddDate(years, months, days)
	}

	var duration time.Duration
	if hours > 0 {
		duration += time.Hour * time.Duration(hours)
	}
	if minutes > 0 {
		duration += time.Minute * time.Duration(minutes)
	}
	if seconds > 0 {
		duration += time.Second * time.Duration(seconds)
	}
	if duration > 0 {
		t = t.Add(duration)
	}

	return t
}
