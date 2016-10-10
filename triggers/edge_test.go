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

package triggers

import (
	"math"
	"testing"

	"github.com/zagrodzki/goscope/scope"
)

var sin = make([]scope.Sample, 10000)

func init() {
	// period of 200 samples and interval 5us, i.e. 1kHz
	// Amplitude +-1V. Total data is 50 cycles.
	for i := range sin {
		sin[i] = scope.Sample(math.Sin(float64(i) / 200 * 2 * math.Pi))
	}
}

func TestTrigger(t *testing.T) {
	in := make(chan scope.Data, 10)
	out := make(chan scope.Data, 10)
	tr := New(in, out)
	// timebase is 400 samples, fits two cycles of the sin.
	// Data should be enough for at least 15 cycles
	// (after 2 cycles captured we might miss one because it starts right
	// after trigger, so we can capture 33 cycles = ~16 samples.
	tr.TimeBase(2 * scope.Millisecond)
	tr.Source("ch1")
	tr.Level(0.3)
	tr.Edge(Falling)

	var sweeps [][]scope.Sample
	var buf []scope.Sample
	var l scope.Duration
	done := make(chan struct{})
	go func() {
		for d := range out {
			buf = append(buf, d.Samples["ch1"]...)
			l += scope.Duration(d.Num) * d.Interval
			if l >= 2*scope.Millisecond {
				l = 0
				sweeps = append(sweeps, buf)
				buf = nil
			}
		}
		if len(buf) > 0 {
			sweeps = append(sweeps, buf)
		}
		done <- struct{}{}
	}()
	for i := 0; i < len(sin)-40; i += 40 {
		t.Logf("Sending samples %d...", i)
		in <- scope.Data{
			Samples: map[scope.ChanID][]scope.Sample{
				"ch1": sin[i : i+40],
			},
			Num:      40,
			Interval: 5 * scope.Microsecond,
		}
	}
	close(in)
	<-done
	if got, want := len(sweeps), 15; got < want {
		t.Fatalf("got %d sweeps, want at least %d", got, want)
	}
	for i, sw := range sweeps[:15] {
		t.Logf("sweep #%d, %d samples", i, len(sw))
		if got, want := len(sw), 400; got < want {
			t.Errorf("sweep #%d: got %d samples, want at least %d", i, got, want)
			continue
		}
		if diff := sw[0] - 0.3; diff > 0.05 || diff < -0.05 {
			t.Errorf("sweep #%d[0]: got %v, want 0.3+-0.05", i, sw[0])
		}
		if s0, s1 := sw[0], sw[1]; s0 <= s1 {
			t.Errorf("sweep #%d[0,1]: got s[0]: %v, s[1]: %v, want s[0] > s[1]", i, s0, s1)
		}
		for j := 0; j < 100; j++ {
			// compare first 100 samples of each trace, they should be almost identical
			if got, want := sweeps[0][j], sw[j]; got-want > 0.01 || got-want < -0.01 {
				t.Errorf("sweep #%d[%d]: got %v, want same as sweep #0[%d] (%v)", i, j, got, j, want)
				break
			}
		}
	}
}
