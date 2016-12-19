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

import (
	"fmt"

	"github.com/zagrodzki/goscope/scope"
)

const (
	paramNameEdge   = "Trigger edge"
	paramNameSource = "Trigger source"
)

// SourceParam is the param controlling trigger signal source channel.
type SourceParam struct {
	ch    scope.ChanID
	avail []string
}

// Name returns the param name for UI.
func (SourceParam) Name() string { return paramNameSource }

// Value returns the current value, name of the source channel.
func (s *SourceParam) Value() string { return string(s.ch) }

// Values returns a list of available source channels.
func (s *SourceParam) Values() []string { return s.avail }

// Set configures a source channel.
func (s *SourceParam) Set(source string) error {
	for _, ch := range s.avail {
		if ch == source {
			s.ch = scope.ChanID(ch)
			return nil
		}
	}
	return fmt.Errorf("Source channel %s is not available. Available sources: %v", s.avail)
}

func newSourceParam(chans []scope.ChanID) *SourceParam {
	avail := make([]string, len(chans))
	for i, ch := range chans {
		avail[i] = string(ch)
	}
	return &SourceParam{
		ch:    chans[0],
		avail: avail,
	}
}

type EdgeParam RisingEdge

// Name returns the param name for UI.
func (EdgeParam) Name() string { return paramNameEdge }

func (e EdgeParam) internal() RisingEdge {
	return RisingEdge(e)
}

// Value returns the current value, type of the triggering edge.
func (e EdgeParam) Value() string {
	switch e.internal() {
	case EdgeRising:
		return "rising"
	case EdgeFalling:
		return "falling"
	}
	return "none"
}

// Values returns a list of edge types.
func (EdgeParam) Values() []string { return []string{"rising", "falling"} }

// Set sets the edge type.
func (e *EdgeParam) Set(v string) error {
	switch v {
	case "rising":
		*e = EdgeParam(EdgeRising)
	case "falling":
		*e = EdgeParam(EdgeFalling)
	default:
		return fmt.Errorf("unknown edge type %q, must be rising or falling", v)
	}
	return nil
}

func newEdgeParam() *EdgeParam {
	ret := EdgeParam(EdgeRising)
	return &ret
}
