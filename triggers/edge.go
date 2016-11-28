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

import "github.com/zagrodzki/goscope/scope"

// RisingEdge represents the trigger edge type, rising or falling
type RisingEdge bool

// RisingEdge values for readability.
const (
	Rising  RisingEdge = true
	Falling RisingEdge = false
)

// Trigger represents a filter running on the data channel, waiting for
// a triggering event and then allowing a set of samples equal to the
// configured timebase.
type Trigger struct {
	source  scope.ChanID
	slope   RisingEdge
	lvl     scope.Voltage
	rec     scope.DataRecorder
	tbCount int
}

// New returns an initialized Trigger.
func New(rec scope.DataRecorder) *Trigger {
	return &Trigger{
		rec: rec,
	}
}

// TimeBase returns the trigger timebase, which is the same as the underlying recorder timebase.
func (t *Trigger) TimeBase() scope.Duration {
	return t.rec.TimeBase()
}

// Reset initializes the recording.
func (t *Trigger) Reset(i scope.Duration, ch <-chan []scope.ChannelData) {
	out := make(chan []scope.ChannelData, 20)
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

func (t *Trigger) run(in <-chan []scope.ChannelData, out chan<- []scope.ChannelData) {
	var trg bool
	var left int
	var last scope.Voltage
	for d := range in {
		var s []scope.Voltage
		for _, ch := range d {
			if ch.ID == t.source {
				s = ch.Samples
				break
			}
		}
		if len(s) == 0 {
			out <- d
			continue
		}
		if !trg {
			for i, v := range s {
				if (last < t.lvl) != (v < t.lvl) && RisingEdge(v >= t.lvl) == t.slope {
					trg = true
					left = t.tbCount
					for ch := range d {
						d[ch].Samples = d[ch].Samples[i:]
					}
					break
				}
				last = v
			}
		}
		if trg {
			if left < len(d[0].Samples) {
				for ch := range d {
					d[ch].Samples = d[ch].Samples[:left]
				}
			}
			left -= len(d[0].Samples)
			if left == 0 {
				trg = false
			}
			out <- d
			last = s[len(s)-1]
		}
	}
	close(out)
}
