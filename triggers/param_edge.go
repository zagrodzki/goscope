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

import "fmt"

const (
	paramNameEdge = "Trigger edge"
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

// Name returns the param name for UI.
func (RisingEdge) Name() string { return paramNameEdge }

// Value returns the current value, type of the triggering edge.
func (e RisingEdge) Value() string {
	switch e {
	case EdgeRising:
		return "rising"
	case EdgeFalling:
		return "falling"
	}
	return "none"
}

// Values returns a list of edge types.
func (RisingEdge) Values() []string { return []string{"rising", "falling"} }

// Set sets the edge type.
func (e *RisingEdge) Set(v string) error {
	switch v {
	case "rising":
		*e = EdgeRising
	case "falling":
		*e = EdgeFalling
	default:
		return fmt.Errorf("unknown edge type %q, must be rising or falling", v)
	}
	return nil
}

func newEdgeParam() *RisingEdge {
	ret := EdgeRising
	return &ret
}
