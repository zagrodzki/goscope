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
	"math/rand"
	"time"

	"github.com/zagrodzki/goscope/scope"
)

type randomChan struct {
	last scope.Voltage
}

var randDiff = func() scope.Voltage {
	return scope.Voltage(rand.NormFloat64() * 0.1)
}

func (randomChan) ID() scope.ChanID { return "random" }
func (ch *randomChan) data(int) []scope.Voltage {
	ret := make([]scope.Voltage, numSamples)
	r := ch.last
	for i := 0; i < numSamples; i++ {
		r = r + randDiff()
		if r > 1.0 {
			r = 1.0
		} else if r < -1.0 {
			r = -1.0
		}
		ret[i] = r
	}
	ch.last = r
	return ret
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
