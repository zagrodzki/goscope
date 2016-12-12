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

	"github.com/pkg/errors"
)

type calData struct {
	max  SampleRate
	data map[rangeID][2]byte
}

func (h *Scope) readCalibrationDataFromDevice() error {
	data := make([]byte, 32)
	n, err := h.dev.Control(controlTypeVendor|controlDirIn, eepromReq, eepromCalibrationOffset, 0, data)
	if err != nil {
		return errors.Wrap(err, "Control(read EEPROM) failed")
	}
	if n != len(data) {
		return fmt.Errorf("Control(read EEPROM): want %d bytes, got %d", len(data), n)
	}
	// alternating bytes for CH1 and CH2. First 16 bytes used for rates <=1Msps, second 16 bytes for >1Msps (<=48Msps)
	// Of 8 bytes per channel, the values are for volt range 0.5, 0.5, 0.5, 1, 2.5, 5, 5, 5. Code uses only offsets 2..5.
	h.calibration = []calData{
		{
			max: 48e6,
			data: map[rangeID][2]byte{
				// data[16..19] are copies of data[20..21]
				voltRange0_5V: [2]byte{data[20], data[21]},
				voltRange1V:   [2]byte{data[22], data[23]},
				voltRange2_5V: [2]byte{data[24], data[25]},
				voltRange5V:   [2]byte{data[26], data[27]},
				// data[28..31] are copies of data[26..27]
			},
		},
		{
			max: 1e6,
			data: map[rangeID][2]byte{
				// data[0..3] are copies of data[4..5]
				voltRange0_5V: [2]byte{data[4], data[5]},
				voltRange1V:   [2]byte{data[6], data[7]},
				voltRange2_5V: [2]byte{data[8], data[9]},
				voltRange5V:   [2]byte{data[10], data[11]},
				// data[12..15] are copies of data[10..11]
			},
		},
	}
	return nil
}

// getCalibration returns values to subtract from samples for each channel, based on current sample rate and measurement ranges.
func (h *Scope) getCalibrationData() [2]float64 {
	var calibration [2]float64
	for _, c := range h.calibration {
		if h.sampleRate <= c.max {
			calibration[ch1Idx] = float64(c.data[h.ch[ch1Idx].voltRange][ch1Idx])
			calibration[ch2Idx] = float64(c.data[h.ch[ch2Idx].voltRange][ch2Idx])
			break
		}
	}
	return calibration
}

// Calibrate performs a calibration of the oscilloscope - it measures the samples for
// ground reference and stores them in the EEPROM on the device.
func (h *Scope) Calibrate() error {
	/*
		TODO(sebek): actually calibrate... procedure in docs/calibration.txt
	*/
	return errors.New("not implemented")
}
