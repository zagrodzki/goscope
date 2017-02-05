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
	"strconv"

	"github.com/zagrodzki/goscope/scope"
)

const (
	paramNameLevel = "level"
)

// Level represents the trigger threshold level. That level combined
// with trigger edge type (rising/falling) determines the trigger condition.
type Level struct {
	v scope.Voltage
}

// Name returns the name of the param.
func (Level) Name() string { return paramNameLevel }

// Value returns the current trigger threshold level.
func (l Level) Value() string { return l.v.String() }

// Set updates the trigger level.
func (l *Level) Set(v string) error {
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fmt.Errorf("ParseFloat(%q): %v", v, err)
	}
	l.v = scope.Voltage(f)
	return nil
}

// Inc increases the trigger level.
// TODO: when Inc is called multiple times in short succession, the change rate should grow.
func (l *Level) Inc() {
	l.v += 0.1
}

// Dec decreases the trigger level.
func (l *Level) Dec() {
	l.v -= 0.1
}

func newLevelParam() *Level {
	return &Level{0}
}
