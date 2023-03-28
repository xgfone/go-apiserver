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

import (
	"fmt"
	"time"
)

// Now is used to customize the time Now.
var Now = time.Now

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

// TryParseTime tries to parse the string value with the layouts in turn to time.Time.
func TryParseTime(loc *time.Location, value string, layouts ...string) (time.Time, error) {
	if len(layouts) == 0 {
		panic("TryParseTime: no time format layouts")
	}

	switch value {
	case "", "0000-00-00 00:00:00", "0000-00-00 00:00:00.000", "0000-00-00 00:00:00.000000":
		return time.Time{}.In(loc), nil
	}

	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, loc); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time '%s'", value)
}

// MustParseTime is the same as time.ParseInLocation, but in turn tries
// to use the layout in layouts to parse value and panics if failed.
//
// If loc is nil, use time.UTC instead.
// If layouts is empty, try to use []string{time.RFC3339Nano, time.DateTime} instead.
func MustParseTime(value string, loc *time.Location, layouts ...string) time.Time {
	if loc == nil {
		loc = time.UTC
	}

	var t time.Time
	var err error
	if len(layouts) == 0 {
		t, err = TryParseTime(loc, value, time.RFC3339Nano, "2006-01-02 15:04:05")
	} else {
		t, err = TryParseTime(loc, value, layouts...)
	}

	if err != nil {
		panic(err)
	}
	return t
}
