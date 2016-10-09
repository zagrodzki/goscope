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
	stuff  chan interface{}
	source scope.ChanID
	slope  RisingEdge
	lvl    scope.Sample
	base   scope.Duration
}

// New returns an initialized Trigger.
func New(in <-chan scope.Data, out chan<- scope.Data) *Trigger {
	tr := &Trigger{
		stuff: make(chan interface{}),
	}
	go tr.run(in, out)
	return tr
}

// Source sets the source for the trigger. If received data doesn't contain
// samples for specified source, the trigger allows all samples without filtering.
func (t *Trigger) Source(id scope.ChanID) {
	t.stuff <- id
}

// Edge configures the type of edge (rising/falling) that is the triggering condition.
func (t *Trigger) Edge(e RisingEdge) {
	t.stuff <- e
}

// Level configures the level that the edge has to cross for the triggering condition.
func (t *Trigger) Level(l scope.Sample) {
	t.stuff <- l
}

// TimeBase sets the trigger timebase - at least that many worth of samples
// will be passed to the output channel after a condition triggers.
func (t *Trigger) TimeBase(d scope.Duration) {
	t.stuff <- d
}

func (t *Trigger) run(in <-chan scope.Data, out chan<- scope.Data) {
	var trg bool
	var lenOut scope.Duration
	var last scope.Sample
	for {
		select {
		case s := <-t.stuff:
			switch v := s.(type) {
			case RisingEdge:
				t.slope = v
			case scope.ChanID:
				t.source = v
			case scope.Sample:
				t.lvl = v
			case scope.Duration:
				t.base = v
			}
		case d, ok := <-in:
			if !ok {
				close(out)
				return
			}
			s, ok := d.Samples[t.source]
			if !ok {
				out <- d
				continue
			}
			if !trg {
				for i, v := range s {
					if (last < t.lvl) != (v < t.lvl) && RisingEdge(v >= t.lvl) == t.slope {
						trg = true
						for ch := range d.Samples {
							d.Samples[ch] = d.Samples[ch][i:]
						}
						d.Num -= i
						break
					}
					last = v
				}
			}
			if trg {
				out <- d
				lenOut += scope.Duration(d.Num) * d.Interval
				if lenOut >= t.base {
					trg = false
					lenOut = 0
				}
				last = s[len(s)-1]
			}
		}
	}
}
