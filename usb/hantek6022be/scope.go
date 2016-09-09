// Package hantek6022be contains a driver for Hantek 6022BE, an inexpensive PC USB oscilloscope.
// The driver uses libusb for communication with the device, based on API
// described in https://github.com/rpcope1/Hantek6022API/blob/master/REVERSE_ENGINEERING.md
package hantek6022be

import (
	"fmt"
	"log"
	"time"

	"bitbucket.org/zagrodzki/goscope/scope"
	"github.com/kylelemons/gousb/usb"
	"github.com/pkg/errors"
)

// Scope is the representation of a Hantek 6022BE USB scope.
type Scope struct {
	dev            *usb.Device
	sampleRate scope.SampleRate
}

// Clear the capture buffer and start sampling.
func (h *Scope) startCapture() error {
	// self.device_handle.controlWrite(0x40, self.TRIGGER_REQUEST, self.TRIGGER_VALUE, self.TRIGGER_INDEX, b'\x01', timeout=timeout)
	_, err := h.dev.Control(0x40, triggerReq, 0, 0, []byte{0x01})
	return errors.Wrap(err, "Control(trigger on) failed")
}

// Stop sampling.
func (h *Scope) stopCapture() error {
	_, err := h.dev.Control(0x40, triggerReq, 0, 0, []byte{0x00})
	return errors.Wrap(err, "Control(trigger off) failed")
}

// String returns a description of the device and it's USB address.
func (h *Scope) String() string {
	return fmt.Sprintf("Hantek 6022BE Oscilloscope at USB bus 0x%x addr 0x%x", h.dev.Bus, h.dev.Address)
}

// GetSampleRate returns the currently configured sample rate.
func (h *Scope) GetSampleRate() scope.SampleRate {
  return h.sampleRate
}

// GetSampleRates returns the list of supported sample rates.
func (*Scope) GetSampleRates() []scope.SampleRate {
  return sampleRates
}

// SetSampleRate sets the desired sample rate {
func (h *Scope) SetSampleRate(s scope.SampleRate) error {
    rate, ok := sampleRateToID[s]
    if !ok {
        return errors.New(fmt.Sprintf("Sample rate %s is not supported by the device, need one of %v", s, sampleRates))
    }
	if _, err := h.dev.Control(0x40, sampleRateReq, 0, 0, rate.data()); err != nil {
		return errors.Wrapf(err, "Control(sample rate %s(%x))", s, rate)
	}
    h.sampleRate = s
    return nil
}

// ReadData reads the sample buffer of the device and returns a SampleResult
// with captured data.
func (h *Scope) ReadData() (map[scope.ChanID][]byte, time.Duration, error) {
	ep, err := h.dev.OpenEndpoint(1, 0, 0, 0x86)
	if err != nil {
		return nil, 0, errors.Wrap(err, "OpenEndpoint(1, 0, 0, 0x86) failed")
	}
	data := make([]byte, 0x800)
    if err := h.startCapture(); err != nil {
      return nil, 0, errors.Wrap(err, "startCapture")
    }
    defer h.stopCapture()
	num, err := ep.Read(data)
	if err != nil {
		return nil, 0, errors.Wrap(err, "ep.Read() failed")
	}
	if num%2 != 0 {
		return nil, 0, errors.New(fmt.Sprintf("Read returned %d bytes of data, expected an even number for 2 channels", num))
	}
	log.Printf("%d bytes read", num)
	ret := make([][]byte, 2)
	ret[0] = make([]byte, num/2)
	ret[1] = make([]byte, num/2)
	for i := 0; i < num; i++ {
		ret[i%2][i/2] = data[i]
	}
	return map[scope.ChanID][]byte{"CH1": ret[0], "CH2": ret[1]}, time.Millisecond, nil
}

// Channels returns a list of channels on the scope, indexed by names matching the channel labels on the device.
func (h *Scope) Channels() map[scope.ChanID]scope.Channel {
	return map[scope.ChanID]scope.Channel{
		"CH1": ch{id: "CH1", osc: h},
		"CH2": ch{id: "CH2", osc: h},
	}
}

// Calibrate performs a calibration of the oscilloscope - it measures the samples for
// ground reference and stores them in the EEPROM on the device.
// TODO(sebek): actually calibrate...
func (h *Scope) Calibrate() error {
	/*
		for _, ch := range h.Channels() {
			if err := ch.SetVoltRange(0.5); err != nil {
				return errors.Wrap(err, fmt.Sprintf("%s.SetVoltRange(0.5)", ch))
			}
		}
		data, err := h.ReadData()
	*/
	return nil
}

// New initializes oscilloscope through the passed USB device.
func New(d *usb.Device) *Scope {
	o := &Scope{dev: d}
	for _, ch := range o.Channels() {
		ch.SetVoltRange(5)
	}
	return o
}

// Close releases the USB device.
func (h *Scope) Close() {
  h.dev.Close()
}

// SupportsUSB will return true if the USB descriptor passed as the argument corresponds to a Hantek 6022BE oscilloscope.
// Used for device autodetection.
func SupportsUSB(d *usb.Descriptor) bool {
	return d.Vendor == hantekVendor && d.Product == hantekProduct
}
