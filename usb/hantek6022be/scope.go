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

// Package hantek6022be contains a driver for Hantek 6022BE, an inexpensive PC USB oscilloscope.
// The driver uses libusb for communication with the device, based on API
// described in https://github.com/rpcope1/Hantek6022API/blob/master/REVERSE_ENGINEERING.md
package hantek6022be

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/zagrodzki/goscope/scope"
	"github.com/zagrodzki/goscope/usb/usbif"
)

// Scope is the representation of a Hantek 6022BE USB scope.
type Scope struct {
	dev         usbif.Device
	sampleRate  SampleRate
	ch          [2]*ch
	stop        chan chan struct{}
	calibration []calData
	rec         scope.DataRecorder
	iso         bool
}

// String returns a description of the device and it's USB address.
func (h *Scope) String() string {
	return fmt.Sprintf("Hantek 6022BE Oscilloscope at USB bus 0x%x addr 0x%x", h.dev.Bus(), h.dev.Address())
}

// setSampleRate sets the desired sample rate {
func (h *Scope) setSampleRate(s SampleRate) error {
	rate, ok := sampleRateToID[s]
	if !ok {
		return errors.Errorf("Sample rate %s is not supported by the device, need one of %v", s, sampleRates)
	}
	if _, err := h.dev.Control(controlTypeVendor, sampleRateReq, 0, 0, rate.data()); err != nil {
		return errors.Wrapf(err, "Control(sample rate %s(%x))", s, rate)
	}
	h.sampleRate = s
	return nil
}

// Attach configures a data recorder for the device.
func (h *Scope) Attach(r scope.DataRecorder) {
	h.rec = r
}
