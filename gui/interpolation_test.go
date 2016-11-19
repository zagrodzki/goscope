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

package gui

import (
	"math"
	"testing"

	"github.com/zagrodzki/goscope/scope"
)

func BenchmarkSincInterpolation(b *testing.B) {
	for _, bc := range []struct {
		name       string
		numSamples int
		interp     Interpolator
	}{
		{
			name:       "non-power of 2",
			numSamples: 1000,
			interp:     SincInterpolator,
		},
		{
			name:       "power of 2",
			numSamples: 1024,
			interp:     SincInterpolator,
		},
		{
			name:       "zero-pad non-power of 2",
			numSamples: 1000,
			interp:     SincZeroPadInterpolator,
		},
		{
			name:       "zero-pad non-power of 2 odd",
			numSamples: 601,
			interp:     SincZeroPadInterpolator,
		},
		{
			name:       "zero-pad power of 2",
			numSamples: 1024,
			interp:     SincZeroPadInterpolator,
		},
	} {
		samples := make([]scope.Voltage, bc.numSamples)
		for i := 0; i < bc.numSamples; i++ {
			samples[i] = scope.Voltage(math.Sin(float64(i) * 4 * math.Pi / float64(bc.numSamples-1)))
		}
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := bc.interp(samples, 2*len(samples))
				if err != nil {
					b.Fatalf("Cannot interpolate: %v", err)
				}
			}
		})
	}
}