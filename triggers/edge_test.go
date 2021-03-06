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
	goodSource    = "signal"
	missingSource = "nonexistent"
)

type fakeDev struct{}

func (fakeDev) String() string            { return "fake" }
func (fakeDev) Channels() []scope.ChanID  { return []scope.ChanID{goodSource, missingSource} }
func (fakeDev) Attach(scope.DataRecorder) {}
func (fakeDev) Start()                    {}
func (fakeDev) Stop()                     {}

func TestTrigger(t *testing.T) {
	// set Auto mode to trigger after 8 samples without the condition.
	defer func(d scope.Duration) { autoDelay = d }(autoDelay)
	autoDelay = 8 * scope.Millisecond

testCases:
	for _, tc := range []struct {
		desc    string
		tbLen   int
		samples [][]scope.Voltage
		level   string
		edge    string
		mode    string
		source  string
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
			level:  "0.3",
			edge:   "falling",
			mode:   "normal",
			source: goodSource,
			want: [][]scope.Voltage{
				{0.2, 0, -0.2, -0.4, -0.6, -0.8, -1, -0.8, -0.6, -0.4},
			},
		},
		{
			desc:  "sawtooth wave, multiple triggers",
			tbLen: 6,
			samples: [][]scope.Voltage{
				{0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7},
				{0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7},
				{0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7},
				{0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7},
			},
			level:  "0.05",
			edge:   "rising",
			mode:   "normal",
			source: goodSource,
			want: [][]scope.Voltage{
				{0.1, 0.2, 0.3, 0.4, 0.5, 0.6},
				{0.1, 0.2, 0.3, 0.4, 0.5, 0.6},
				{0.1, 0.2, 0.3, 0.4, 0.5, 0.6},
				{0.1, 0.2, 0.3, 0.4, 0.5, 0.6},
			},
		},
		{
			desc:  "sawtooth wave, single mode - only one trigger",
			tbLen: 6,
			samples: [][]scope.Voltage{
				{0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7},
				{0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7},
				{0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7},
				{0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7},
			},
			level:  "0.05",
			edge:   "rising",
			mode:   "single",
			source: goodSource,
			want: [][]scope.Voltage{
				{0.1, 0.2, 0.3, 0.4, 0.5, 0.6},
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
			level:  "0.25",
			edge:   "rising",
			mode:   "normal",
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
			level:  "-1",
			edge:   "falling",
			mode:   "normal",
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
			level:  "-1",
			edge:   "rising",
			mode:   "normal",
			source: goodSource,
			want:   nil,
		},
		{
			desc:  "never reaches the threshold, auto mode",
			tbLen: 6,
			samples: [][]scope.Voltage{
				{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8},
				{0.81, 0.82, 0.83, 0.84, 0.85, 0.86, 0.87, 0.88},
			},
			level:  "1",
			edge:   "rising",
			mode:   "auto",
			source: goodSource,
			want: [][]scope.Voltage{
				{0.81, 0.82, 0.83, 0.84, 0.85, 0.86},
			},
		},
		{
			desc:  "constant samples at the threshold, rising edge",
			tbLen: 4,
			samples: [][]scope.Voltage{
				{1, 1, 1, 1, 1, 1, 1},
				{1, 1, 1, 1, 1},
				{1, 1, 1, 1, 1, 1, 1, 1},
			},
			level:  "1",
			edge:   "rising",
			mode:   "normal",
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
			level:  "1",
			edge:   "falling",
			mode:   "normal",
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
			level:  "0",
			edge:   "rising",
			mode:   "normal",
			source: goodSource,
			want: [][]scope.Voltage{
				{1, 1, 1, 1, 1, 1, 1, 1},
			},
		},
		{
			desc:  "source not present in data",
			tbLen: 8,
			samples: [][]scope.Voltage{
				{-10, -9, -8, -7, -6, -5, -4, -3, -2, -1},
				{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			},
			level:  "0",
			edge:   "rising",
			mode:   "normal",
			source: missingSource,
			want: [][]scope.Voltage{
				{-10, -9, -8, -7, -6, -5, -4, -3, -2, -1},
				{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			},
		},
		{
			desc:  "trigger disabled through mode 'none'",
			tbLen: 8,
			samples: [][]scope.Voltage{
				{-10, -9, -8, -7, -6, -5, -4, -3, -2, -1},
				{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			},
			level:  "0",
			edge:   "rising",
			mode:   "none",
			source: goodSource,
			want: [][]scope.Voltage{
				{-10, -9, -8, -7, -6, -5, -4, -3, -2, -1},
				{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			},
		},
	} {
		buf := testutil.NewBufferRecorder(scope.Duration(tc.tbLen) * scope.Millisecond)
		tr := New(&fakeDev{})
		tr.Attach(buf)
		for _, p := range tr.TriggerParams() {
			var err error
			switch p.Name() {
			case paramNameEdge:
				err = p.Set(tc.edge)
			case paramNameMode:
				err = p.Set(tc.mode)
			case paramNameLevel:
				err = p.Set(tc.level)
			case paramNameSource:
				err = p.Set(tc.source)
			}
			if err != nil {
				t.Errorf("%s: TriggerParams[%q].Set: %v", tc.desc, p.Name(), err)
				continue testCases
			}
		}

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
			t.Errorf("%s: got %d sweeps, want %d. Full sweeps:\n%v", tc.desc, got, want, sweeps)
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
				if got[j] != want[j] {
					t.Errorf("%s: sweep #%d[%d]: got %v, want %v. Full sweeps:\nGot: %v\nWant: %v", tc.desc, i, j, got[j], want[j], got, want)
					break compareSweeps
				}
			}
		}
	}
}
