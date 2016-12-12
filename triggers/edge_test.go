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
	"github.com/zagrodzki/goscope/testutil"
)

var sin = make([]scope.Voltage, 10000)

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
		mode    Mode
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
			edge:   EdgeFalling,
			mode:   ModeNormal,
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
			edge:   EdgeRising,
			mode:   ModeNormal,
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
			edge:   EdgeFalling,
			mode:   ModeNormal,
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
			edge:   EdgeRising,
			mode:   ModeNormal,
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
			edge:   EdgeRising,
			mode:   ModeNormal,
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
			edge:   EdgeFalling,
			mode:   ModeNormal,
			source: goodSource,
			want:   nil,
		},
		{
			desc:  "crosses the threshold slowly",
			tbLen: 8,
			samples: [][]scope.Voltage{
				{-1, -1, -1, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				{1, 1, 1, 1, 1, 1, 1},
			},
			level:  0,
			edge:   EdgeRising,
			mode:   ModeNormal,
			source: goodSource,
			want: [][]scope.Voltage{
				{1, 1, 1, 1, 1, 1, 1, 1},
			},
		},
	} {
		buf := testutil.NewBufferRecorder(scope.Duration(tc.tbLen) * scope.Millisecond)
		tr := New(buf)
		tr.Source(tc.source)
		tr.Level(tc.level)
		tr.Edge(tc.edge)
		tr.Mode(tc.mode)

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
		sweeps, _ := buf.Wait()

		if got, want := len(sweeps), len(tc.want); got != want {
			t.Errorf("%s: got %d sweeps, want %d", tc.desc, got, want)
			continue
		}
	compareSweeps:
		for i, got := range sweeps {
			want := tc.want[i]
			if len(got) != len(want) {
				t.Errorf("%s: sweep #%d: got %d samples, want %d. Full sweeps:\nGot %v\nWant: %v", tc.desc, i, len(got), len(want), got, want)
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
