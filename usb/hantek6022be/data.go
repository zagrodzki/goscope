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

import "github.com/zagrodzki/goscope/scope"

const (
	hantekVendor  = 0x4b5
	hantekProduct = 0x6022

	ch1VoltRangeReq uint8 = 0xe0
	ch2VoltRangeReq uint8 = 0xe1

	sampleRateReq uint8 = 0xe2
	triggerReq    uint8 = 0xe3
	// 0xe4 is supported only with opensource firmware.
	// numChReq      uint8 = 0xe4

	eepromReq               uint8  = 0xa2
	eepromCalibrationOffset uint16 = 0x08
	eepromCalibrationLen    int    = 0x20

	bulkConfig    uint8 = 1
	bulkInterface uint8 = 0
	bulkAlt       uint8 = 0
	bulkEP        uint8 = 0x86

	// constants from libusb, defined by USB spec.
	controlTypeMask   uint8 = 0x60
	controlTypeVendor uint8 = 0x40
	controlDirMask    uint8 = 0x80
	controlDirOut     uint8 = 0x00
	controlDirIn      uint8 = 0x80

	ch1ID   scope.ChanID = "CH1"
	ch2ID   scope.ChanID = "CH2"
	ch1Idx               = 0
	ch2Idx               = 1
	numChan              = 2
)

type rateID uint8

func (s rateID) data() []byte {
	return []byte{byte(s)}
}

var (
	sampleRates = []scope.SampleRate{100e3, 200e3, 500e3, 1e6, 4e6, 8e6, 16e6}
	// Rates 24e6, 30e6, 48e6 are available, but USB bus speed is limited to
	// 60MB/s in theory, and to 40ishMB/s in practice. With 48e6 samples per
	// channel per second the transfer rate would have to be 90MB/s+ to
	// sustain the read. Not enough bus throughput means the device
	// captures samples at 48Msps into the 2kB buffer in the device,
	// and then pauses while FIFO is full. That's generally not useful,
	// as there is no triggering in hardware and there's no continuous data
	// stream to the host.
	// We might still use 48Msps rate for calibration, because the signal
	// level during calibration is expected to be constant.
	// TODO(sebek): with custom firmware 6022BE can report samples from
	// only one channel, so 30e6 or 48e6 might be feasible. Not supported yet.
	sampleRateToID = map[scope.SampleRate]rateID{
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
	sampleIDToRate = make(map[rateID]scope.SampleRate)
)

type rangeID uint8

// usb packet data for range request
func (v rangeID) data() []byte {
	return []byte{byte(v)}
}

var (
	voltRanges    = []scope.VoltRange{0.5, 1, 2.5, 5}
	voltRangeToID = map[scope.VoltRange]rangeID{
		5:   0x01,
		2.5: 0x02,
		1:   0x05,
		0.5: 0x0a,
	}
	voltIDToRange = make(map[rangeID]scope.VoltRange)
)

func init() {
	for r, id := range voltRangeToID {
		voltIDToRange[id] = r
	}
	for s, id := range sampleRateToID {
		sampleIDToRate[id] = s
	}
}
