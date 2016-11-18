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

package hantek6022be

import (
	"github.com/pkg/errors"
	"github.com/zagrodzki/goscope/scope"
)

type ch struct {
	id        scope.ChanID
	osc       *Scope
	voltRange uint8
}

func (c ch) ID() scope.ChanID { return c.id }
func (c *ch) setVoltRange(v uint8) error {
	var req uint8
	switch c.id {
	case "CH1":
		req = ch1VoltRangeReq
	case "CH2":
		req = ch2VoltRangeReq
	}
	if _, err := c.osc.dev.Control(controlTypeVendor, req, 0, 0, []uint8{v}); err != nil {
		return errors.Wrapf(err, "Control(voltage range %x)", v)
	}
	c.voltRange = v
	return nil
}

// Channels returns a list of channel names on the scope, names matching the channel labels on the device.
func (h *Scope) Channels() []scope.ChanID {
	return []scope.ChanID{ch1ID, ch2ID}
}
