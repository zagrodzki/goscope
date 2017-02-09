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
	"log"

	"github.com/pkg/errors"
	"github.com/zagrodzki/goscope/triggers"
	"github.com/zagrodzki/goscope/usb/usbif"
)

// New initializes oscilloscope through the passed USB device.
func New(d usbif.Device) (*triggers.Trigger, error) {
	o := &Scope{dev: d}
	for _, c := range d.Configs() {
		if c.Config != isoConfig {
			continue
		}
		for _, intf := range c.Interfaces {
			if intf.Number != isoInterface {
				continue
			}
			for _, s := range intf.Setups {
				if s.Alternate != isoAlt {
					continue
				}
				for _, ep := range s.Endpoints {
					if ep.Address == isoEP && ep.Attributes&transferTypeMask == transferTypeIso {
						o.iso = true
					}
				}
			}
		}
	}
	if !o.iso {
		log.Print("Using bulk transfers, suitable for original firmware. Device performs better with isochronous transfers, available with alternative modded firmware. See http://foo for details.")
	}
	o.ch = [2]*ch{
		{id: "CH1", osc: o},
		{id: "CH2", osc: o},
	}
	for _, ch := range o.ch {
		ch.setVoltRange(voltRange5V)
	}
	o.setSampleRate(1e6)
	if err := o.readCalibrationDataFromDevice(); err != nil {
		return nil, errors.Wrap(err, "readCalibration")
	}
	return triggers.New(o), nil
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
