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

package scope

import "fmt"

// ChanID represents the ID of a probe channel on a scope.
type ChanID string

// Voltage represents a single sample value, in Volts
type Voltage float64

// String returns a string representation of the value.
func (v Voltage) String() string {
	return fmt.Sprintf("%.4f", v)
}

// ChannelData represents a sequence of samples for a single channel.
type ChannelData struct {
	ID      ChanID
	Samples []Voltage
}

// Data represents a set of samples collected from the scope.
type Data struct {
	// Channels contains the sample data per channel.
	Channels []ChannelData
	// Num is the sample count.
	Num int
	// Interval indicates the time period between samples.
	Interval Duration
	// Error is not nil if an error occured during data collection. Values of Samples and Interval are unspecified if Error is not nil.
	Error error
}
