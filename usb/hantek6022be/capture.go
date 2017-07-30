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
	readLen := uint64(h.numChan) * uint64(tb) / uint64(h.sampleInterval)
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
	calibration [2]float64
	scale       [2]scope.Voltage
}

var sampleBuf []byte

// get samples from USB and send processed to channel.
func (h *Scope) getSamples(ep reader, p captureParams, ch chan<- []scope.ChannelData) error {
	num, err := ep.Read(sampleBuf)
	if err != nil {
		return errors.Wrap(err, "Read")
	}
	if num%h.numChan != 0 {
		return errors.Errorf("Read returned %d bytes of data, expected a number divisible by %d for %d channels", num, h.numChan, h.numChan)
	}
	var samples [2][]scope.Voltage
	for i := 0; i < h.numChan; i++ {
		samples[i] = make([]scope.Voltage, num/h.numChan)
	}
	cal := [2]calVoltage{
		h.cal.voltage[0][h.ch[0].voltRange],
		h.cal.voltage[1][h.ch[1].voltRange],
	}
	for i, ch := 0, 0; i < num; i, ch = i+1, (ch+1)%h.numChan {
		samples[ch][i/h.numChan] = scope.Voltage(sampleBuf[i])*cal[ch].interval + cal[ch].min
	}
	ch <- []scope.ChannelData{
		{ID: ch1ID, Samples: samples[ch1Idx]},
		{ID: ch2ID, Samples: samples[ch2Idx]},
	}[:h.numChan]

	return nil
}

// Start starts processing of USB data.
func (h *Scope) Start() {
	// buffer for 20 samples, don't keep the data collection hanging.
	ret := make(chan []scope.ChannelData, 2)
	h.rec.Reset(h.sampleInterval, ret)
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

	// TODO(sebek): later move the offset calculation here.
	params := captureParams{}

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
