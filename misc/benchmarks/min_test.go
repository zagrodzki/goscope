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

package benchmark

import (
	"math"
	"testing"
)

func minSimple(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func minMath(a, b int) int {
	return int(math.Min(float64(a), float64(b)))
}

func BenchmarkMin(b *testing.B) {
	var out int
	for _, bc := range []struct {
		name string
		min  func(int, int) int
	}{
		{"simple", minSimple},
		{"math", minMath},
	} {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				out = bc.min(3, 11)
			}
		})
		b.Logf("%v", out)
	}
}
