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

// Clear the capture buffer and start sampling.
func (h *Scope) startCapture() error {
	if h.stop != nil {
		return nil
	}
	if _, err := h.dev.Control(controlTypeVendor, triggerReq, 0, 0, []byte{0x01}); err != nil {
		return errors.Wrap(err, "Control(trigger on) failed")
	}
	h.stop = make(chan chan struct{}, 1)
	// HT6022BE has only 2kB buffer onboard. At 16Msps it takes about 60us to fill it up.
	// Request as much data as possible in one go, that way the host does not have
	// to spend time going back and forth between sending commands and receiving data,
	// but just keeps reading data packets.
	// But cap the ep.Read latency below 1/10th of a second to ensure high-ish refresh rate.
	tb := scope.Millisecond * 100
	if recTB := h.rec.TimeBase(); tb > 8*recTB {
		tb = 8 * recTB
	}
	readLen := (uint64(h.sampleRate) * uint64(h.numChan) * uint64(tb)) / uint64(scope.Second)
	switch h.iso {
	case true:
		// round up to 3072, the maximum ISO transfer packet size.
		if readLen%3072 != 0 {
			readLen = 3072 * (readLen/3072 + 1)
		}
	case false:
		// round up to nearest 512B
		if readLen%512 != 0 {
			readLen = 512 * (readLen/512 + 1)
		}
	}
	sampleBuf = make([]byte, readLen)
	return nil
}

// Stop sampling.
func (h *Scope) stopCapture() error {
	_, err := h.dev.Control(controlTypeVendor, triggerReq, 0, 0, []byte{0x00})
	return errors.Wrap(err, "Control(trigger off) failed")
}

type reader interface {
	Read([]byte) (int, error)
}

type captureParams struct {
	translateSample [2][256]scope.Voltage
}

var sampleBuf []byte

// get samples from USB and send processed to channel.
func (h *Scope) getSamples(ep reader, p *captureParams, ch chan<- []scope.ChannelData) error {
	num, err := ep.Read(sampleBuf)
	if err != nil {
		return errors.Wrap(err, "Read")
	}
	if num%h.numChan != 0 {
		return errors.Errorf("Read returned %d bytes of data, expected a number divisible by %d for %d channels", num, h.numChan, h.numChan)
	}
	samples := make([]scope.Voltage, num)
	perChan := num / h.numChan
	for ch := 0; ch < h.numChan; ch++ {
		trans := p.translateSample[ch]
		for in, out := ch, ch*perChan; in < num; in, out = in+h.numChan, out+1 {
			samples[out] = trans[sampleBuf[in]]
		}
	}
	ret := make([]scope.ChannelData, h.numChan)
	for idx, id := range []scope.ChanID{ch1ID, ch2ID}[:h.numChan] {
		ret[idx] = scope.ChannelData{ID: id, Samples: samples[idx*perChan : (idx+1)*perChan]}
	}
	ch <- ret
	return nil
}

// Start starts processing of USB data.
func (h *Scope) Start() {
	// buffer for 20 samples, don't keep the data collection hanging.
	ret := make(chan []scope.ChannelData, 2)
	h.rec.Reset(h.sampleRate.Interval(), ret)
	usbCfg := bulkConfig
	usbIf := bulkInterface
	usbAlt := bulkAlt
	usbEP := bulkEP
	if h.iso {
		usbCfg = isoConfig
		usbIf = isoInterface
		usbAlt = isoAlt
		usbEP = isoEP
	}
	ep, err := h.dev.OpenEndpoint(usbCfg, usbIf, usbAlt, usbEP)
	if err != nil {
		h.rec.Error(errors.Wrap(err, "OpenEndpoint"))
		close(ret)
		return
	}

	params := &captureParams{}
	calibration := h.getCalibrationData()
	scale := [2]scope.Voltage{
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
		h.ch[ch1Idx].voltRange.volts() / 123,
		h.ch[ch2Idx].voltRange.volts() / 123,
	}
	for idx := range params.translateSample {
		for i := 0; i < 256; i++ {
			params.translateSample[idx][i] = scope.Voltage(float64(i)-calibration[idx]) * scale[idx]
		}
	}

	if err := h.startCapture(); err != nil {
		h.rec.Error(errors.Wrap(err, "startCapture"))
		close(ret)
		return
	}

	go func() {
		defer close(ret)
		for {
			select {
			case stopped := <-h.stop:
				h.stopCapture()
				close(stopped)
				return
			default:
				if err := h.getSamples(ep, params, ret); err != nil {
					h.rec.Error(errors.Wrap(err, "getSamples"))
					h.stopCapture()
					return
				}
			}
		}
	}()
}

// Stop halts the data capture goroutine.
func (h *Scope) Stop() {
	ret := make(chan struct{})
	h.stop <- ret
	<-ret
}
