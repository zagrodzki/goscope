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
	"testing"

	"github.com/zagrodzki/goscope/scope"
)

var sin = make([]scope.Voltage, 10000)

type dataRec struct {
	tb     int
	i      scope.Duration
	sweeps [][]scope.Voltage
	err    error
	done   chan struct{}
}

func (r *dataRec) TimeBase() scope.Duration {
	return scope.Millisecond * scope.Duration(r.tb)
}

func (r *dataRec) Reset(i scope.Duration, ch <-chan []scope.ChannelData) {
	r.i = i
	r.done = make(chan struct{})
	tbCount := int(r.TimeBase() / r.i)
	var buf []scope.Voltage
	var l int
	go func() {
		for d := range ch {
			l += len(d[0].Samples)
			buf = append(buf, d[0].Samples...)
			if l >= tbCount {
				l = 0
				r.sweeps = append(r.sweeps, buf)
				buf = nil
			}
		}
		if len(buf) > 0 {
			r.sweeps = append(r.sweeps, buf)
		}
		close(r.done)
	}()
}

func (r *dataRec) Error(err error) {
	r.err = err
}

const (
	goodSource    = scope.ChanID("signal")
	missingSource = scope.ChanID("nonexistent")
)

func TestTrigger(t *testing.T) {
	for _, tc := range []struct {
		desc    string
		tbLen   int
		samples [][]scope.Voltage
		level   scope.Voltage
		edge    RisingEdge
		source  scope.ChanID
		want    [][]scope.Voltage
	}{
		{
			desc:  "triangle wave, simple trigger on falling edge",
			tbLen: 10,
			samples: [][]scope.Voltage{
				{0, 0.2, 0.4, 0.8},
				{1.0, 0.8, 0.6, 0.4},
				{0.2, 0, -0.2, -0.4},
				{-0.6, -0.8, -1, -0.8},
				{-0.6, -0.4, -0.2, 0},
			},
			level:  0.3,
			edge:   Falling,
			source: goodSource,
			want: [][]scope.Voltage{
				{0.2, 0, -0.2, -0.4, -0.6, -0.8, -1, -0.8, -0.6, -0.4},
			},
		},
		{
			desc:  "first sample above threshold, triggers only on actual crossing",
			tbLen: 8,
			samples: [][]scope.Voltage{
				{0.5, 0.7, 0.9, 1.1},
				{-0.1, 0.1, 0.2, 0.3},
				{0.4, 0.5, 0.6, 0.7},
				{0.8, 0.9, 1.0, 1.1},
			},
			level:  0.25,
			edge:   Rising,
			source: goodSource,
			want: [][]scope.Voltage{
				{0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
			},
		},
		{
			desc:  "never reaches the threshold, falling edge",
			tbLen: 8,
			samples: [][]scope.Voltage{
				{1, 1, 1, 1, 1, 1, 1, 1},
				{1, 1, 1, 1, 1, 1, 1, 1},
				{1, 1, 1, 1, 1, 1, 1, 1},
			},
			level:  -1,
			edge:   Falling,
			source: goodSource,
			want:   nil,
		},
		{
			desc:  "never reaches the threshold, rising edge",
			tbLen: 8,
			samples: [][]scope.Voltage{
				{1, 1, 1, 1, 1, 1, 1, 1},
				{1, 1, 1, 1, 1, 1, 1, 1},
				{1, 1, 1, 1, 1, 1, 1, 1},
			},
			level:  -1,
			edge:   Rising,
			source: goodSource,
			want:   nil,
		},
		{
			desc:  "constant samples at the threshold, rising edge",
			tbLen: 4,
			samples: [][]scope.Voltage{
				{1, 1, 1, 1, 1, 1, 1},
				{1, 1, 1, 1, 1},
				{1, 1, 1, 1, 1, 1, 1, 1},
			},
			level:  1,
			edge:   Rising,
			source: goodSource,
			want:   nil,
		},
		{
			desc:  "constant samples at the threshold, falling edge",
			tbLen: 4,
			samples: [][]scope.Voltage{
				{1, 1, 1, 1, 1, 1, 1},
				{1, 1, 1, 1, 1},
				{1, 1, 1, 1, 1, 1, 1, 1},
			},
			level:  1,
			edge:   Falling,
			source: goodSource,
			want:   nil,
		},
	} {
		buf := &dataRec{tb: tc.tbLen}
		tr := New(buf)
		tr.Source(tc.source)
		tr.Level(tc.level)
		tr.Edge(tc.edge)

		in := make(chan []scope.ChannelData, 10)
		tr.Reset(scope.Millisecond, in)

		for _, v := range tc.samples {
			in <- []scope.ChannelData{
				{
					ID:      goodSource,
					Samples: v,
				},
			}
		}
		close(in)
		<-buf.done

		if got, want := len(buf.sweeps), len(tc.want); got != want {
			t.Errorf("%s: got %d sweeps, want %d", tc.desc, got, want)
			continue
		}
	compareSweeps:
		for i, got := range buf.sweeps {
			want := tc.want[i]
			if len(got) != len(want) {
				t.Errorf("%s: sweep #%d: got %d samples, want %d. Full sweeps:\nGot %v\nWant: %v", tc.desc, len(got), len(want), got, want)
				break
			}
			for j := 0; j < len(got); j++ {
				t.Logf("Sweep %d, #%d, %v vs %v", i, j, got[j], want[j])
				if got[j] != want[j] {
					t.Errorf("%s: sweep #%d[%d]: got %v, want %v. Full sweeps:\nGot: %v\nWant: %v", tc.desc, i, j, got[j], want[j], got, want)
					break compareSweeps
				}
			}
		}
	}
}
