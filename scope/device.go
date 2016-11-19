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

// Device represents a connected sampling device (e.g. USB oscilloscope).
type Device interface {
	// String returns a description of the device. It should be specific enough
	// to allow the user to identify the physical device that this value
	// represents.
	String() string

	// Channels returns list of available channel IDs.
	Channels() []ChanID

	// StartSampling starts reading data off the device.
	// This interface assumes all channels on a single Device are sampled at the
	// same rate and return the same number of samples for every run.
	// Stop function should be called by the user when device should stop sampling.
	// After calling stop, user should keep reading from data channel until
	// that channel is closed.
	// If the device encounters an error, that error will be returned within the
	// channel (as Data.Error). The channel may be closed by the device after
	// encountering an error.
	StartSampling() (data <-chan Data, stop func(), err error)
}
