//  Copyright 2016 The goscope Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package scope

import (
	"fmt"
	"time"
)

// Duration represents an interval between samples, measured in femtoseconds.
// Duration is very similar to time.Duration, except time.Duration does not have
// sufficient precision to describe intervals between samples at high rates.
// In gigasamples per second ranges, intervals are in the order of picoseconds.
type Duration uint64

func (d Duration) String() string {
	var sfx string
	var div float64 = 1
	switch {
	case d == 0:
		return "0s"
	case d >= 1e12:
		return time.Duration(d / 1e6).String()
	case d < 1000:
		sfx = "fs"
	case d < 1e6:
		sfx = "ps"
		div = 1e3
	case d < 1e9:
		sfx = "ns"
		div = 1e6
	default:
		sfx = "µs"
		d = d / 1e3
		div = 1e6
	}
	ret := fmt.Sprintf("%f", float64(d)/div)
	for ret[len(ret)-1] == '0' {
		ret = ret[:len(ret)-1]
	}
	if ret[len(ret)-1] == '.' {
		ret = ret[:len(ret)-1]
	}
	return fmt.Sprintf("%s%s", ret, sfx)
}

// Common durations.
const (
	Femtosecond Duration = 1
	Picosecond  Duration = 1e3
	Nanosecond  Duration = 1e6
	Microsecond Duration = 1e9
	Millisecond Duration = 1e12
	Second      Duration = 1e15
)

// DurationFromNano converts a time.Duration (duration in nanoseconds) to
// an equivalent scope.Duration.
func DurationFromNano(d time.Duration) Duration {
	return Duration(d) * Nanosecond
}
