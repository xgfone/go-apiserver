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

package helper

import "time"

// Now is used to customize the time Now.
var Now = time.Now

// NowLocal returns the now local time.
func NowLocal() time.Time { return Now().Local() }

// NowUTC returns the now UTC time.
func NowUTC() time.Time { return Now().UTC() }

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

// MustParseTime is the same as time.ParseInLocation, but panics if failed.
//
// If layout is empty, use time.RFC3339 instead.
// If loc is nil, use time.UTC instead.
func MustParseTime(layout, value string, loc *time.Location) time.Time {
	if layout == "" {
		layout = time.RFC3339
	}

	if loc == nil {
		loc = time.UTC
	}

	t, err := time.ParseInLocation(layout, value, loc)
	if err != nil {
		panic(err)
	}

	return t
}
