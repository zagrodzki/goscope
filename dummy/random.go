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

type randomChan struct{}

var randLast scope.Sample

var randDiff = func() scope.Sample {
	return scope.Sample(rand.NormFloat64() * 0.1)
}

func (randomChan) ID() scope.ChanID                   { return "random" }
func (randomChan) GetVoltRange() scope.VoltRange      { return 1 }
func (randomChan) GetVoltRanges() []scope.VoltRange   { return []scope.VoltRange{1} }
func (randomChan) SetVoltRange(scope.VoltRange) error { return nil }
func (r randomChan) data() []scope.Sample {
	ret := make([]scope.Sample, numSamples)
	for i := 0; i < numSamples; i++ {
		randLast = randLast + randDiff()
		if randLast > 1.0 {
			randLast = 1.0
		} else if randLast < -1.0 {
			randLast = -1.0
		}
		ret[i] = randLast
	}
	return ret
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
