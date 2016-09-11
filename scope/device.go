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

// Package scope defines an abstract interface for a digital oscilloscope or other
// similar capture device.
package scope

import (
	"fmt"
	"time"
)

// SampleRate represents a Device sampling frequency in samples/second.
type SampleRate int

// String returns a human-readable representation of sampling rate.
func (s SampleRate) String() string {
	return fmt.Sprintf("%s samples/s", fmtVal(float64(s)))
}

// Sample represents a single sample value, in Volts
type Sample float64

// Device represents a connected sampling device (e.g. USB oscilloscope).
type Device interface {
	// String returns a description of the device. It should be specific enough
	// to allow the user to identify the physical device that this value
	// represents.
	String() string

	// Channels returns a map of Channels indexed by their IDs. Channel can be used
	// to configure parameters related to a single capture source.
	Channels() map[ChanID]Channel

	// ReadData asks the device for a trace.
	// This interface assumes all channels on a single Device are sampled at the
	// same rate and return the same number of samples for every run.
	ReadData() (map[ChanID][]Sample, time.Duration, error)

	// GetSampleRate returns the currently configured sample rate.
	GetSampleRate() SampleRate

	// GetSampleRates returns a slice of sample rates available on this device.
	GetSampleRates() []SampleRate
}
