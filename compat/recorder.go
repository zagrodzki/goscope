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

package compat

import "github.com/zagrodzki/goscope/scope"

// Recorder implements the scope.Recorder interface to allow attaching
// devices. At the same time it exposes the same sort of API that was used
// previously, with a channel for reading sample data.
type Recorder struct {
	TB       scope.Duration
	Data     chan scope.Data
	interval scope.Duration
}

// TimeBase returns the configured timebase.
func (g *Recorder) TimeBase() scope.Duration {
	return g.TB
}

// Reset initializes the recorder. The Data channel is initialized only after Reset.
func (g *Recorder) Reset(i scope.Duration) {
	g.Stop()
	g.interval = i
	g.Data = make(chan scope.Data, 1)
}

// Record writes a set of samples to the recorder. That data is passed onto the Data channel.
func (g *Recorder) Record(d []scope.ChannelData) {
	if len(d) == 0 {
		return
	}
	g.Data <- scope.Data{
		Channels: d,
		Num:      len(d[0].Samples),
		Interval: g.interval,
	}
}

// Stop is called when no more data will be recorder before next Reset.
func (g *Recorder) Stop() {
	if g.Data != nil {
		close(g.Data)
		g.Data = nil
	}
}

// Error reports an error to the recorder. Error is passed onto the Data channel.
func (g *Recorder) Error(err error) {
	g.Data <- scope.Data{
		Error: err,
	}
}
