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

package testutil

import "github.com/zagrodzki/goscope/scope"

// Recorder is a buffer that implements the scope.DataRecorder interface.
// Users should call Wait() to get the data that was recorded in the buffer.
type BufferRecorder struct {
	tb     scope.Duration
	i      scope.Duration
	sweeps [][]scope.Voltage
	err    error
	done   chan struct{}
}

// NewRecorder creates a new test data recorder with timebase equal to tb.
func NewBufferRecorder(tb scope.Duration) *BufferRecorder {
	return &BufferRecorder{
		tb: tb,
	}
}

// TimeBase returns the configured timebase (sweep length) of the recorder.
func (r *BufferRecorder) TimeBase() scope.Duration {
	return r.tb
}

// Reset prepares a new recording with sample interval i, reading samples from ch.
func (r *BufferRecorder) Reset(i scope.Duration, ch <-chan []scope.ChannelData) {
	r.i = i
	r.done = make(chan struct{})
	tbCount := int(r.tb / r.i)
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

// Error reports an acquisition error to the data recorder.
func (r *BufferRecorder) Error(err error) {
	r.err = err
}

// Wait waits until the source finishes writing data to the recorder.
// It then returns recorded data and last error reported (if any).
func (r *BufferRecorder) Wait() ([][]scope.Voltage, error) {
	<-r.done
	return r.sweeps, r.err
}
