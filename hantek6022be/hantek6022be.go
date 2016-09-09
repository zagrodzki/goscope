package hantek6022be

import (
	"fmt"
	"log"
	"time"

	"github.com/kylelemons/gousb/usb"
	"github.com/pkg/errors"
	"zagrodzki.net/gohantek/oscilloscope"
)

const (
	myVendor  = 0x4b5
	myProduct = 0x6022

	sampleRateReq uint8 = 0xe2
	triggerReq    uint8 = 0xe3
	numChReq      uint8 = 0xe4
)

type sampleRate uint8

var sampleRates = map[int]sampleRate{
	100e3: 0x0a,
	200e3: 0x14,
	500e3: 0x32,
	1e6:   0x01,
	4e6:   0x04,
	8e6:   0x08,
	16e6:  0x10,
	24e6:  0x30,
}

type osc struct {
	dev *usb.Device
    sampleDuration time.Duration
}

func (h *osc) String() string {
	return fmt.Sprintf("Hantek 6022BE Oscilloscope at USB bus 0x%x addr 0x%x", h.dev.Bus, h.dev.Address)
}

func (h *osc) StartCapture() error {
	// self.device_handle.controlWrite(0x40, self.TRIGGER_REQUEST, self.TRIGGER_VALUE, self.TRIGGER_INDEX, b'\x01', timeout=timeout)
	_, err := h.dev.Control(0x40, triggerReq, 0, 0, []byte{0x01})
	return errors.Wrap(err, "Control(trigger on) failed")
}

func (h *osc) StopCapture() error {
	_, err := h.dev.Control(0x40, triggerReq, 0, 0, []byte{0x00})
	return errors.Wrap(err, "Control(trigger off) failed")
}

func (h *osc) ReadData() (map[oscilloscope.ChanID][]byte, time.Duration, error) {
	ep, err := h.dev.OpenEndpoint(1, 0, 0, 0x86)
	if err != nil {
		return nil, 0, errors.Wrap(err, "OpenEndpoint(1, 0, 0, 0x86) failed")
	}
	data := make([]byte, 0x800)
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
	return map[oscilloscope.ChanID][]byte{"CH1": ret[0], "CH2": ret[1]}, time.Millisecond, nil
}

func (h *osc) Channels() map[oscilloscope.ChanID]oscilloscope.Channel {
	return map[oscilloscope.ChanID]oscilloscope.Channel{
		"CH1": ch{"CH1", h},
		"CH2": ch{"CH2", h},
	}
}

func (h *osc) Calibrate() error {
	for _, ch := range h.Channels() {
		if err := ch.SetVoltRange(0.5); err != nil {
			return errors.Wrap(err, fmt.Sprintf("%s.SetVoltRange(0.5)"), ch)
		}
	}
	if _, err := h.StartCapture(); err != nil {
		return errors.Wrap(err, "StartCapture")
	}
	data, err := ReadData()
}

func newOsc(d *usb.Device) *osc {
	return &osc{dev: d}
}

func New(d *usb.Device) oscilloscope.Device {
	o := newOsc(d)
    for _, ch := range o.Channels() {
      ch.SetVoltRange(5)
    }
    return o
}

func Supports(d *usb.Descriptor) bool {
	return d.Vendor == myVendor && d.Product == myProduct
}
