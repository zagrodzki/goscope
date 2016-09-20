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
	"github.com/zagrodzki/goscope/usb/usbif"
)

// New initializes oscilloscope through the passed USB device.
func New(d usbif.Device) (*Scope, error) {
	o := &Scope{dev: d}
	o.ch = [2]*ch{
		{id: "CH1", osc: o},
		{id: "CH2", osc: o},
	}
	for _, ch := range o.ch {
		ch.SetVoltRange(5)
	}
	o.SetSampleRate(1e6)
	if err := o.readCalibrationDataFromDevice(); err != nil {
		return nil, errors.Wrap(err, "readCalibration")
	}
	return o, nil
}

// Close releases the USB device.
func (h *Scope) Close() {
	h.dev.Close()
}

// SupportsUSB will return true if the USB descriptor passed as the argument corresponds to a Hantek 6022BE oscilloscope.
// Used for device autodetection.
func SupportsUSB(d *usbif.Desc) bool {
	return d.Vendor == hantekVendor && d.Product == hantekProduct
}
