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
	"flag"
	"fmt"
	"log"

	"github.com/kylelemons/gousb/usb"
	"github.com/pkg/errors"
	"github.com/zagrodzki/goscope/triggers"
	"github.com/zagrodzki/goscope/usb/usbif"
)

var (
	sampleRate = flag.Uint("sample_rate_ksps", 1000, "Sample rate, in Ksps. Supported: 100, 200, 500, 1000, 4000, 8000, 16000")
	voltRange  = flag.Uint("measurement_range", 1, "Measurement range. 1: +-5V, 2: +-2.5V, 5: +-1V, 10: +-0.5V")
	disableCH2 = flag.Bool("disable_ch2", false, "When set, CH2 is disabled, leaving more USB bandwidth for CH1. Allows use of 16/24Msps")
	forceBulk  = flag.Bool("force_bulk", false, "When set, bulk transfers are used even when isochronous transfers are available.")
)

// New initializes oscilloscope through the passed USB device.
func New(d usbif.Device) (*triggers.Trigger, error) {
	o := &Scope{dev: d, numChan: 2}
	if err := d.SetConfig(usbConfig); err != nil {
		return nil, fmt.Errorf("SetConfig(%d): %v", usbConfig, err)
	}
	c, err := d.Config()
	if err != nil {
		return nil, fmt.Errorf("Config(): %v", err)
	}
	if usbInterface >= len(c.Interfaces) {
		return nil, fmt.Errorf("device %s does not have interface %d", d, usbInterface)
	}
	intf := c.Interfaces[usbInterface]
	if isoAlt < len(intf.AltSettings) {
		alt := intf.AltSettings[isoAlt]
		for _, ep := range alt.Endpoints {
			if ep.Number == isoEP && ep.TransferType == usb.TransferTypeIsochronous {
				o.customFW = true
				o.forceBulk = *forceBulk
				break
			}
		}
	}
	if !o.customFW {
		log.Print(`Using bulk transfers, the only option with the original firmware.
Device performs better with isochronous transfers,
available with alternative open-source firmware.
See http://foo for details.`)
	}
	o.ch = [2]*ch{
		{id: "CH1", osc: o},
		{id: "CH2", osc: o},
	}
	for _, ch := range o.ch {
		if err := ch.setVoltRange(rangeID(*voltRange)); err != nil {
			return nil, fmt.Errorf("setVoltRange(%s, %d): %v", ch.id, *voltRange, err)
		}
	}
	if o.customFW {
		numChan := 2
		if *disableCH2 {
			numChan = 1
		}
		if err := o.setNumChan(numChan); err != nil {
			return nil, fmt.Errorf("setNumChan(%d): %v", numChan, err)
		}
	}
	if err := o.setSampleRate(SampleRate(*sampleRate * 1000)); err != nil {
		return nil, fmt.Errorf("setSampleRate(%d): %v", *sampleRate, err)
	}
	// TODO(sebek): add reading calibration data from a file.
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
