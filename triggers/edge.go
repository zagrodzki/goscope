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
	"time"

	"github.com/zagrodzki/goscope/scope"
)

// autoDelay controls the delay for triggering without condition in ModeAuto.
// Used in tests.
var autoDelay = 500 * scope.Millisecond

// Trigger represents a filter running on the data channel, waiting for
// a triggering event and then allowing a set of samples equal to the
// configured timebase.
// Trigger implements both scope.Device interface (used by UI)
// and scope.DataRecorder interface (used by underlying device).
type Trigger struct {
	scope.Device
	source   *Source
	slope    *RisingEdge
	lvl      *Level
	rec      scope.DataRecorder
	interval scope.Duration
	tbCount  int
	mode     *Mode
}

// New returns an initialized Trigger.
func New(dev scope.Device) *Trigger {
	return &Trigger{
		Device: dev,
		mode:   newModeParam(),
		slope:  newEdgeParam(),
		lvl:    newLevelParam(),
		source: newSourceParam(dev.Channels()),
	}
}

// Attach configures the trigger to write filtered data to rec.
func (t *Trigger) Attach(rec scope.DataRecorder) {
	t.rec = rec
	t.Device.Attach(t)
}

// TimeBase returns the trigger timebase, which is the same as the underlying recorder timebase.
func (t *Trigger) TimeBase() scope.Duration {
	return t.rec.TimeBase()
}

// Reset initializes the recording.
func (t *Trigger) Reset(i scope.Duration, ch <-chan []scope.ChannelData) {
	if *t.mode == ModeNone {
		t.rec.Reset(i, ch)
		return
	}
	out := make(chan []scope.ChannelData, 2)
	t.interval = i
	t.tbCount = int(t.rec.TimeBase() / i)
	t.rec.Reset(i, out)
	go t.run(ch, out)
}

// Error passes the error down to the underlying recorder.
func (t *Trigger) Error(err error) {
	t.rec.Error(err)
}

// TriggerParams returns the trigger params.
func (t *Trigger) TriggerParams() []scope.Param {
	return []scope.Param{
		t.slope,
		t.mode,
		t.lvl,
		t.source,
	}
}

type thresholdState int

const (
	belowThreshold        thresholdState = -1
	unknownThresholdState thresholdState = 0
	aboveThreshold        thresholdState = 1
)

func edgeType(prevState, newState thresholdState) RisingEdge {
	switch {
	case prevState == newState:
		return EdgeNone
	case prevState < newState:
		return EdgeRising
	case prevState > newState:
		return EdgeFalling
	}
	return EdgeNone
}

type slice struct {
	begin int
	end   int
}

func (t *Trigger) run(in <-chan []scope.ChannelData, out chan<- []scope.ChannelData) {
	var left, source, ignored int
	var trg, scanned, found bool
	var newState, prevState thresholdState
	var lastTrg time.Time
	var sensitivity scope.Voltage = 0.05
	maxIgnored := int(autoDelay / t.interval)
	slope := *t.slope
	mode := *t.mode
	lvl := t.lvl.v
	for d := range in {
		if !scanned {
			scanned = true
			for i := range d {
				if d[i].ID == t.source.ch {
					source = i
					found = true
					break
				}
			}
		}
		if !found {
			out <- d
			continue
		}
		num := len(d[source].Samples)
		// slices keeps indices of the samples that should be pushed out
		var outSlices []slice
		var curSlice slice
		for i, v := range d[source].Samples {
			switch {
			case v > lvl+sensitivity:
				newState = aboveThreshold
			case v < lvl-sensitivity:
				newState = belowThreshold
			}
			// if the previous state was uninitialized, do not trigger.
			// Once state is initialized, it's always either above or below, never unknown.
			if newState != prevState && prevState == unknownThresholdState {
				prevState = newState
			}
			// newState > prevState means we moved from below threshold to above threshold, i.e. rising slope.
			if !trg {
				switch {
				// mode single and triggered once already. Don't trigger.
				case mode == ModeSingle && !lastTrg.IsZero():
				// crossed the threshold
				case edgeType(prevState, newState) == slope:
					trg = true
				// mode auto and time elapsed since last trigger.
				case mode == ModeAuto && ignored >= maxIgnored:
					trg = true
				}
				if trg {
					lastTrg = time.Now()
					left = t.tbCount
					curSlice.begin = i
					ignored = 0
				}
			}
			if trg {
				curSlice.end = i + 1
				left--
				if left == 0 {
					outSlices = append(outSlices, curSlice)
					curSlice = slice{}
					trg = false
				}
			} else {
				ignored++
			}
			prevState = newState
			num--
		}
		if trg {
			outSlices = append(outSlices, curSlice)
		}
		// flush samples
		if len(outSlices) > 0 {
			for _, b := range outSlices {
				chunk := make([]scope.ChannelData, len(d))
				for ch := range d {
					chunk[ch].ID = d[ch].ID
					chunk[ch].Samples = d[ch].Samples[b.begin:b.end]
				}
				out <- chunk
			}
			outSlices = outSlices[:0]
		}
	}
	close(out)
}
