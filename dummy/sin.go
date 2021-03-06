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

	"github.com/zagrodzki/goscope/scope"
)

type sinChan struct{}

func (sinChan) ID() scope.ChanID { return "sin" }
func (ch sinChan) data(offset int) []scope.Voltage {
	ret := make([]scope.Voltage, numSamples)
	inc := 60 * math.Pi
	flof := float64(offset)
	var phase float64
	for i := 0; i < numSamples; i++ {
		ret[i] = scope.Voltage(math.Sin(phase/numSamples + flof))
		phase += inc
	}
	return ret
}
