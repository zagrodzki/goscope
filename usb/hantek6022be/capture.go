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
	"github.com/zagrodzki/goscope/scope"
)

// Clear the capture buffer and start sampling.
func (h *Scope) startCapture() error {
	if h.stop != nil {
		return nil
	}
	if _, err := h.dev.Control(controlTypeVendor, triggerReq, 0, 0, []byte{0x01}); err != nil {
		return errors.Wrap(err, "Control(trigger on) failed")
	}
	h.stop = make(chan struct{}, 1)
	// HT6022BE has only 2kB buffer onboard. At 16Msps it takes about 60us to fill it up.
	// Request as much data as possible in one go, that way the host does not have
	// to spend time going back and forth between sending commands and receiving data,
	// but just keeps reading data packets.
	// But cap the ep.Read latency below 1/10th of a second to ensure high-ish refresh rate.
	readLen := int(h.sampleRate * numChan / 10)
	// round up to nearest 512B
	if readLen%512 != 0 {
		readLen = 512 * (readLen/512 + 1)
	}
	sampleBuf = make([]byte, readLen)
	return nil
}

// Stop sampling.
func (h *Scope) stopCapture() error {
	if h.stop == nil {
		return nil
	}
	h.stop <- struct{}{}
	h.stop = nil
	_, err := h.dev.Control(controlTypeVendor, triggerReq, 0, 0, []byte{0x00})
	return errors.Wrap(err, "Control(trigger off) failed")
}

type reader interface {
	Read([]byte) (int, error)
}

type captureParams struct {
	calibration [2]float64
	scale       [2]float64
}

var sampleBuf []byte

// get samples from USB and send processed to channel.
func (h *Scope) getSamples(ep reader, p captureParams, ch chan<- scope.Data) error {
	num, err := ep.Read(sampleBuf)
	if num != len(sampleBuf) {
		log.Printf("Read %d bytes, buffer lenght %d", num, len(sampleBuf))
	}
	if err != nil {
		return errors.Wrap(err, "Read")
	}
	if num%numChan != 0 {
		return errors.Errorf("Read returned %d bytes of data, expected an even number for 2 channels", num)
	}
	samples := [2][]scope.Sample{make([]scope.Sample, num/numChan), make([]scope.Sample, num/numChan)}
	for i := 0; i < num; i++ {
		samples[i%numChan][i/numChan] = scope.Sample((float64(sampleBuf[i]) - p.calibration[i%numChan]) * p.scale[i%numChan])
	}
	ch <- scope.Data{
		Samples: map[scope.ChanID][]scope.Sample{
			ch1ID: samples[ch1Idx],
			ch2ID: samples[ch2Idx],
		},
		Num:      num / numChan,
		Interval: h.sampleRate.Interval(),
	}
	return nil
}

// StartSampling starts processing of USB data.
func (h *Scope) StartSampling() (<-chan scope.Data, func(), error) {
	ep, err := h.dev.OpenEndpoint(bulkConfig, bulkInterface, bulkAlt, bulkEP)
	if err != nil {
		return nil, nil, errors.Wrap(err, "OpenEndpoint")
	}

	params := captureParams{
		calibration: h.getCalibrationData(),
		scale: [2]float64{
			// TODO(sebek): /123 is a very poor approximation.
			// The actual channel measurement range is not as specified (0.5/1/2.5/5), but quite a bit off.
			// For example, a quick test with a calibrated power supply shows for my HT6022BE:
			// CH1: 5V: -5.19..5.08, 2.5V: -2.67..2.58, 1V: -1.04..1.02, 0.5V: -0.529..0.523
			// CH2: 5V: -5.52..4.66, 2.5V: -2.8..2.38, 1V: -1.1..0.94, 0.5V: -0.562..0.481
			// With calibration values of:
			// CH1: 5V: 128, 2.5V: 128, 1V: 127, 0.5V: 126
			// CH2: 5V: 135, 2.5V: 135, 1V: 135, 0.5V: 135
			// That suggests that the actual measured extremes are not following declared measurement range in a linear way.
			// The bare minimum would be to store additional byte per channel/range in calibration data:
			// what's the change in measured byte value corresponding to a change equal to half of measurement range.
			// There are 16 unused bytes in calibration data, this would require 8 bytes.
			// I.e. if measurement range is +-5V, how much needs to be added to a certain measurement value to represent 5V higher voltage.
			// Ideally this would be measured from a reference voltage equal to half of measurement extremum (e.g. 2.5V) and confirmed
			// by measuring the same reference in reverse polarity.
			// But it's also possible to ask the user what reference voltage was used.
			// It should also be possible to set zero for calibration by finding the middle point between reverse polarity measurements.
			// Note that probe at 1x might have a different range than at 10x.
			// For PP80B, data sheet specifies 2% tolerance, so switching between 1x/10x might introduce up to 4% difference in reaadings.
			float64(h.ch[ch1Idx].voltRange) / 123,
			float64(h.ch[ch2Idx].voltRange) / 123,
		},
	}
	// buffer for 20 samples, don't keep the data collection hanging.
	ret := make(chan scope.Data, 20)

	if err := h.startCapture(); err != nil {
		return nil, nil, errors.Wrap(err, "startCapture")
	}

	go func() {
		defer h.stopCapture()
		for {
			select {
			case <-h.stop:
				close(ret)
				return
			default:
				if err := h.getSamples(ep, params, ret); err != nil {
					ret <- scope.Data{Error: errors.Wrap(err, "getSamples")}
					if err := h.stopCapture(); err != nil {
						ret <- scope.Data{Error: errors.Wrap(err, "stopCapture")}
					}
					return
				}
			}
		}
	}()

	return ret, func() { h.stopCapture() }, nil
}
