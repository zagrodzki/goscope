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

// RisingEdge represents the trigger edge type, rising or falling
type RisingEdge int

const (
	// EdgeNone represents unknown edge type.
	EdgeNone RisingEdge = iota
	// EdgeRising represents a signal crossing from below to above the threshold.
	EdgeRising
	// EdgeFalling represents a signal crossing from above to below the threshold.
	EdgeFalling
)

// Mode represents the triggering mode, see comments in the constants below.
type Mode int

const (
	// ModeNone means unknown mode.
	ModeNone = iota
	// ModeSingle means trigger once and never again.
	ModeSingle
	// ModeNormal means trigger on every condition, but don't ever trigger
	// without the condition present. Might result in long intervals where
	// data is discarded.
	ModeNormal
	// ModeAuto is like ModeNormal, but will also trigger after some time
	// (currently hardcoded to 0.5s) has passed without the trigger.
	ModeAuto
)

// autoDelay controls the delay for triggering without condition in ModeAuto.
// Used in tests.
var autoDelay = 500 * scope.Millisecond

// Trigger represents a filter running on the data channel, waiting for
// a triggering event and then allowing a set of samples equal to the
// configured timebase.
type Trigger struct {
	source   scope.ChanID
	slope    RisingEdge
	lvl      scope.Voltage
	rec      scope.DataRecorder
	interval scope.Duration
	tbCount  int
	mode     Mode
}

// New returns an initialized Trigger.
func New(rec scope.DataRecorder) *Trigger {
	return &Trigger{
		rec:  rec,
		mode: ModeAuto,
	}
}

// TimeBase returns the trigger timebase, which is the same as the underlying recorder timebase.
func (t *Trigger) TimeBase() scope.Duration {
	return t.rec.TimeBase()
}

// Reset initializes the recording.
func (t *Trigger) Reset(i scope.Duration, ch <-chan []scope.ChannelData) {
	out := make(chan []scope.ChannelData, 20)
	t.interval = i
	t.tbCount = int(t.rec.TimeBase() / i)
	t.rec.Reset(i, out)
	go t.run(ch, out)
}

// Error passes the error down to the underlying recorder.
func (t *Trigger) Error(err error) {
	t.rec.Error(err)
}

// Source sets the source for the trigger. If received data doesn't contain
// samples for specified source, the trigger allows all samples without filtering.
func (t *Trigger) Source(id scope.ChanID) {
	t.source = id
}

// Edge configures the type of edge (rising/falling) that is the triggering condition.
func (t *Trigger) Edge(e RisingEdge) {
	t.slope = e
}

// Level configures the level that the edge has to cross for the triggering condition.
func (t *Trigger) Level(l scope.Voltage) {
	t.lvl = l
}

// Mode sets the trigger mode.
func (t *Trigger) Mode(m Mode) {
	t.mode = m
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
	maxIgnored := int(autoDelay / t.interval)
	for d := range in {
		if !scanned {
			scanned = true
			for i := range d {
				if d[i].ID == t.source {
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
			case v > t.lvl:
				newState = aboveThreshold
			case v < t.lvl:
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
				case t.mode == ModeSingle && !lastTrg.IsZero():
				// crossed the threshold
				case edgeType(prevState, newState) == t.slope:
					trg = true
				// mode auto and time elapsed since last trigger.
				case t.mode == ModeAuto && ignored >= maxIgnored:
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
