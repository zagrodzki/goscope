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

package dummy

import (
	"math"
	"testing"

	"github.com/zagrodzki/goscope/scope"
)

func TestRandom(t *testing.T) {
	ch := &randomChan{}
	data := ch.data(0)
	min, max := data[0], data[0]
	avgDiff := data[len(data)-1] / scope.Voltage(len(data))
	var last scope.Voltage
	var stdDev scope.Voltage
	last = 0
	for _, s := range data {
		switch {
		case s > max:
			max = s
		case s < min:
			min = s
		}
		diff := s - last
		last = s
		stdDev += (diff - avgDiff) * (diff - avgDiff)
	}
	stdDev /= scope.Voltage(len(data))

	if got, min, max := math.Sqrt(float64(stdDev)), 0.05, 0.15; got < min || got > max {
		t.Errorf("sample difference stddev squared: expected between %v and %v, got %v", min, max, got)
	}
	if want := scope.Voltage(1); max != want {
		t.Errorf("maximal sample value in the data set: %v, want less equal than %v", max, want)
	}
	if want := scope.Voltage(-1); min != want {
		t.Errorf("minimal sample value in the data set: %v, want greater equal than %v", min, want)
	}
}
