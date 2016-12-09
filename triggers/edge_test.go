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
	"errors"
	"math"
	"testing"

	"github.com/zagrodzki/goscope/scope"
)

var sin = make([]scope.Voltage, 10000)

func init() {
	// period of 200 samples and interval 5us, i.e. 1kHz
	// Amplitude +-1V. Total data is 50 cycles.
	for i := range sin {
		sin[i] = scope.Voltage(math.Sin(float64(i) / 200 * 2 * math.Pi))
	}
}

type dataRec struct {
	i      scope.Duration
	sweeps [][]scope.Voltage
	err    error
	done   chan struct{}
}

func (*dataRec) TimeBase() scope.Duration {
	return 2 * scope.Millisecond
}

func (r *dataRec) Reset(i scope.Duration, ch <-chan []scope.ChannelData) {
	r.i = i
	r.done = make(chan struct{}, 1)
	tbCount := int(r.TimeBase() / r.i)
	var buf []scope.Voltage
	var l int
	go func() {
		for d := range ch {
			for _, ch := range d {
				if ch.ID == scope.ChanID("ch1") {
					buf = append(buf, ch.Samples...)
					l += len(ch.Samples)
					break
				}
			}
			if l >= tbCount {
				l = 0
				r.sweeps = append(r.sweeps, buf)
				buf = nil
			}
		}
		if len(buf) > 0 {
			r.sweeps = append(r.sweeps, buf)
		}
		r.done <- struct{}{}
	}()
}

func (r *dataRec) Error(err error) {
	r.err = err
}

func TestTrigger(t *testing.T) {
	in := make(chan []scope.ChannelData, 10)
	buf := &dataRec{}
	tr := New(buf)
	// timebase is 400 samples, fits two cycles of the sin.
	// Data should be enough for at least 15 cycles
	// (after 2 cycles captured we might miss one because it starts right
	// after trigger, so we can capture 33 cycles = ~16 samples.
	tr.Source("ch1")
	tr.Level(0.3)
	tr.Edge(Falling)

	tr.Reset(5*scope.Microsecond, in)
	for i := 0; i < len(sin)-40; i += 40 {
		in <- []scope.ChannelData{
			{
				ID:      "ch1",
				Samples: sin[i : i+40],
			},
		}
	}
	tr.Error(errors.New("some error"))
	close(in)
	<-buf.done

	if got, want := buf.i, 5*scope.Microsecond; got != want {
		t.Errorf("recorder interval: got %s, want %s", got, want)
	}
	if buf.err == nil {
		t.Error("recorder error: expected some error, got nil")
	}
	if got, want := len(buf.sweeps), 15; got < want {
		t.Fatalf("got %d sweeps, want at least %d", got, want)
	}
	for i, sw := range buf.sweeps[:15] {
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
			if got, want := sw[j], buf.sweeps[0][j]; got-want > 0.01 || got-want < -0.01 {
				t.Errorf("sweep #%d[%d]: got %v, want same as sweep #0[%d] (%v+-0.01 )", i, j, got, j, want)
				break
			}
		}
	}
}
