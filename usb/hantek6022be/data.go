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
	"fmt"
	"log"

	"github.com/zagrodzki/goscope/scope"
)

const (
	hantekVendor  = 0x4b5
	hantekProduct = 0x6022

	ch1VoltRangeReq uint8 = 0xe0
	ch2VoltRangeReq uint8 = 0xe1

	sampleRateReq uint8 = 0xe2
	triggerReq    uint8 = 0xe3
	numChReq      uint8 = 0xe4 // custom firmware only

	eepromReq               uint8  = 0xa2
	eepromCalibrationOffset uint16 = 0x08
	eepromCalibrationLen    int    = 0x20

	bulkConfig    uint8 = 1
	bulkInterface uint8 = 0
	bulkAlt       uint8 = 0
	bulkEP        uint8 = 0x86

	isoConfig    uint8 = 1
	isoInterface uint8 = 0
	isoAlt       uint8 = 1
	isoEP        uint8 = 0x82

	// constants from libusb, defined by USB spec.
	controlTypeMask   uint8 = 0x60
	controlTypeVendor uint8 = 0x40
	controlDirMask    uint8 = 0x80
	controlDirOut     uint8 = 0x00
	controlDirIn      uint8 = 0x80
	transferTypeMask  uint8 = 0x03
	transferTypeIso   uint8 = 0x01

	ch1ID  scope.ChanID = "CH1"
	ch2ID  scope.ChanID = "CH2"
	ch1Idx              = 0
	ch2Idx              = 1
)

type rateID uint8

func (s rateID) data() []byte {
	return []byte{byte(s)}
}

var (
	sampleRates = []SampleRate{100e3, 200e3, 500e3, 1e6, 4e6, 8e6, 16e6, 24e6}
	// Rates 30e6, 48e6 are available, but USB bus speed is limited to
	// 60MB/s in theory, and to 40ishMB/s in practice. With 48e6 samples per
	// channel per second the transfer rate would have to be 90MB/s+ to
	// sustain the read. Not enough bus throughput means the device
	// captures samples at 48Msps into the 2kB buffer in the device,
	// and then pauses while FIFO is full. That's generally not useful,
	// as there is no triggering in hardware and there's no continuous data
	// stream to the host.
	// We might still use 48Msps rate for calibration, because the signal
	// level during calibration is expected to be constant.
	//
	// With custom firmware and using isochronous mode, the scope can use
	// a max/guaranteed bandwidth of 24MBps and allows the use of a single
	// channel, allowing up to 24Msps.
	sampleRateToID = map[SampleRate]rateID{
		100e3: 0x0a,
		200e3: 0x14,
		500e3: 0x32,
		1e6:   0x01,
		4e6:   0x04,
		8e6:   0x08,
		16e6:  0x10,
		24e6:  0x18,
		30e6:  0x1e,
		48e6:  0x30,
	}
	sampleIDToRate = make(map[rateID]SampleRate)
)

type rangeID uint8

// usb packet data for range request
func (v rangeID) data() []byte {
	return []byte{byte(v)}
}

func (v rangeID) volts() scope.Voltage {
	switch v {
	case voltRange5V:
		return 5.0
	case voltRange2_5V:
		return 2.5
	case voltRange1V:
		return 1.0
	case voltRange0_5V:
		return 0.5
	default:
		log.Fatalf("Unknown voltage range ID: %v", v)
	}
	return 0
}

const (
	voltRange5V   rangeID = 0x01
	voltRange2_5V rangeID = 0x02
	voltRange1V   rangeID = 0x05
	voltRange0_5V rangeID = 0x0a
)

// SampleRate represents a Device sampling frequency in samples/second.
type SampleRate int

// String returns a human-readable representation of sampling rate.
func (s SampleRate) String() string {
	return fmt.Sprintf("%s samples/s", fmtVal(float64(s)))
}

// Interval returns an interval between two samples for given rate.
func (s SampleRate) Interval() scope.Duration {
	return scope.Second / scope.Duration(s)
}

func init() {
	for s, id := range sampleRateToID {
		sampleIDToRate[id] = s
	}
}
