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
	paramNameMode = "mode"
)

// Mode represents the triggering mode, see comments in the constants below.
type Mode int

const (
	// ModeNone means trigger disabled.
	ModeNone = Mode(iota)
	// ModeSingle means trigger once and never again.
	ModeSingle
	// ModeNormal means trigger on every condition, but don't ever trigger
	// without the condition present. Might result in long intervals where
	// data is discarded.
	ModeNormal
	// ModeAuto is like ModeNormal, but will also trigger after some time
	// (currently hardcoded to 0.5s) has passed without the trigger.
	ModeAuto
)

// Name returns the name of the parameter for the UI.
func (Mode) Name() string { return paramNameMode }

// Value returns the string representation of the current mode.
func (m Mode) Value() string {
	switch m {
	case ModeSingle:
		return "single"
	case ModeNormal:
		return "normal"
	case ModeAuto:
		return "auto"
	}
	return "none"
}

// Values returns a list of available modes.
func (Mode) Values() []string {
	return []string{"single", "normal", "auto"}
}

// Set sets the mode.
func (m *Mode) Set(v string) error {
	switch v {
	case "none":
		*m = ModeNone
	case "single":
		*m = ModeSingle
	case "normal":
		*m = ModeNormal
	case "auto":
		*m = ModeAuto
	default:
		return fmt.Errorf("unknown trigger mode %q, must be single, normal or auto", v)
	}
	return nil
}

func newModeParam() *Mode {
	return new(Mode)
}
