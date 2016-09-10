package hantek6022be

import "bitbucket.org/zagrodzki/goscope/scope"

const (
	hantekVendor  = 0x4b5
	hantekProduct = 0x6022

	ch1VoltRangeReq uint8 = 0xe0
	ch2VoltRangeReq uint8 = 0xe1

	sampleRateReq uint8 = 0xe2
	triggerReq    uint8 = 0xe3
	numChReq      uint8 = 0xe4
)

type rateID uint8

func (s rateID) data() []byte {
	return []byte{byte(s)}
}

var (
	sampleRates    = []scope.SampleRate{100e3, 200e3, 500e3, 1e6, 4e6, 8e6, 16e6, 24e6}
	sampleRateToID = map[scope.SampleRate]rateID{
		100e3: 0x0a,
		200e3: 0x14,
		500e3: 0x32,
		1e6:   0x01,
		4e6:   0x04,
		8e6:   0x08,
		16e6:  0x10,
		24e6:  0x30,
	}
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
)
