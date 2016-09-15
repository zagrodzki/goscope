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

// VoltRange represents a measure range in Volts.
type VoltRange float64

// String returns a human-readable representation of measurement range.
func (v VoltRange) String() string {
	return fmt.Sprintf("+-%fV", v)
}

// Channel represents the probe channel on a scope.
type Channel interface {
	// ID returns the channel ID
	ID() ChanID

	// GetVoltRange returns the currently configured measurement range.
	GetVoltRange() VoltRange

	// GetVoltRanges returns a slice with available ranges that can be passed to SetVoltRange.
	GetVoltRanges() []VoltRange

	// SetVoltRange adjusts the sensitivity
	SetVoltRange(VoltRange) error
}
