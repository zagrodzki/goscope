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

// Param is a control element of the Device - a button, knob, switch on the scope front panel.
type Param interface {
	// Name returns the name of the param, intended for the UI.
	Name() string
	// Value returns the current value displayed as string, for the UI.
	Value() string
	// Set sets a new value of the param. If the value is invalid, a non-nil error is returned.
	Set(string) error
}

// SelectParam represents anything that would be a button or a discrete
// knob/slider on a physical device. This control represents discrete settings
// that don’t have an ordering - e.g. name, type etc. or a range with
// few discrete values (e.g. time base length).
// For ranges with many values (unsuitable for a drop-down list),
// a RangeParam is better.
type SelectParam interface {
	Param
	// Values returns the list of available values to choose from.
	Values() []string
}

// RangeParam represents a continuous control knob/slider.
// The speed of change may accelerate when the user
// keeps increasing the value at a quick steady pace.
type RangeParam interface {
	Param
	// Inc increases the param setting.
	Inc() string
	// Dec decreases the param setting.
	Dec() string
}
