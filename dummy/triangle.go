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

import "github.com/zagrodzki/goscope/scope"

type triangleChan struct{}

func (triangleChan) ID() scope.ChanID { return "triangle" }
func (ch triangleChan) data(offset int) []scope.Voltage {
	ret := make([]scope.Voltage, numSamples)
	for i := 0; i < numSamples; i++ {
		if (i+offset)%40 < 20 {
			ret[i] = scope.Voltage(float64((i+offset)%20-10) / 10)
		} else {
			ret[i] = scope.Voltage(float64(30-(i+offset)%40) / 10)
		}
	}
	return ret
}
