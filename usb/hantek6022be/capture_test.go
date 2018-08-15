//  Copyright 2017 The goscope Authors
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

package hantek6022be

import (
	"testing"

	"github.com/zagrodzki/goscope/scope"
)

var (
	calibration = []float64{128, 128}
	scale       = []scope.Voltage{0.04, 0.04}
	numChan     = 2
)

func getSamplesSingleMake(buf []byte) [][]scope.Voltage {
	num := len(buf)
	samples := make([]scope.Voltage, num)
	step := num / numChan
	for ch := 0; ch < numChan; ch++ {
		for src, dst := ch, ch*step; src < num; src, dst = src+numChan, dst+1 {
			samples[dst] = 99
		}
	}
	return [][]scope.Voltage{samples[:step], samples[step:]}
}

func getSamplesTwoMakes(buf []byte) [][]scope.Voltage {
	num := len(buf)
	samples := [][]scope.Voltage{
		make([]scope.Voltage, num/numChan),
		make([]scope.Voltage, num/numChan),
	}
	for ch := 0; ch < numChan; ch++ {
		for src, dst := ch, 0; src < num; src, dst = src+numChan, dst+1 {
			samples[ch][dst] = 99
		}
	}
	return samples
}

var voltLookup [][]scope.Voltage

func init() {
	voltLookup = make([][]scope.Voltage, numChan)
	for ch := range voltLookup {
		voltLookup[ch] = make([]scope.Voltage, 256)
		for i := 0; i < 256; i++ {
			voltLookup[ch][i] = scope.Voltage(float64(i)-calibration[ch]) * scale[ch]
		}
	}
}

func getSamplesFastLookupTwo(buf []byte) [][]scope.Voltage {
	num := len(buf)
	samples := make([][]scope.Voltage, numChan)
	for ch := 0; ch < numChan; ch++ {
		samples[ch] = make([]scope.Voltage, num/numChan)
		base := voltLookup[ch]
		out := samples[ch]
		for dst, src := 0, ch; src < num; dst, src = dst+1, src+numChan {
			out[dst] = base[buf[src]]
		}
	}
	return samples
}

func getSamplesFastLookupOne(buf []byte) [][]scope.Voltage {
	num := len(buf)
	samples := make([][]scope.Voltage, numChan)
	for ch := range samples {
		samples[ch] = make([]scope.Voltage, num/numChan)
	}
	ch, dst := 0, 0
	for i := 0; i < num; i++ {
		samples[ch][dst] = voltLookup[ch][buf[i]]
		ch++
		if ch >= numChan {
			ch = 0
			dst++
		}
	}
	return samples
}

func BenchmarkGetSamples(b *testing.B) {
	buf := make([]byte, 20000)
	for i := range buf {
		buf[i] = 100 + byte(i%60)
	}
	var out [][]scope.Voltage
	for _, tc := range []struct {
		name string
		f    func([]byte) [][]scope.Voltage
	}{
		{
			name: "single make",
			f:    getSamplesSingleMake,
		},
		{
			name: "two makes",
			f:    getSamplesTwoMakes,
		},
		{
			name: "fast lookup one loop",
			f:    getSamplesFastLookupOne,
		},
		{
			name: "fast lookup two loops",
			f:    getSamplesFastLookupTwo,
		},
	} {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				out = tc.f(buf)
			}
		})
	}
	// Make govet happy.
	_ = out
}
