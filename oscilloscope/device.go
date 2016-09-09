package oscilloscope

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

// ChanID represents the ID of a probe channel on a scope.
type ChanID string

// VoltRange represents a measure range in Volts.
type VoltRange float64

// String returns a human-readable representation of measurement range.
func (v VoltRange) String() {
	return fmt.Sprintf("+-%fV", v)
}

// Channel represents the probe channel on a scope.
type Channel interface {
	// ID returns the channel ID
	ID() ChanID

	// GetVoltRanges returns a slice with available ranges that can be passed to SetVoltRange.
	GetVoltRanges() []VoltRange
	// SetVoltRange adjusts the sensitivity
	SetVoltRange(VoltRange) error
}

func fmtVal(v float64) string {
	av := math.Abs(v)
	sfx := ""
	switch {
	case av >= 1e9:
		v /= 1e9
		sfx = "G"
	case av >= 1e6:
		v /= 1e6
		sfx = "M"
	case av >= 1e3:
		v /= 1e3
		sfx = "K"
	}
	ret := strconv.FormatFloat(v, 'f', 3, 64)
	for ret[len(ret)-1] == '0' {
		ret = ret[:len(ret)-1]
	}
	if ret[len(ret)-1] == '.' {
		ret = ret[:len(ret)-1]
	}
	return fmt.Sprintf("%s%s", ret, sfx)
}

// SampleRate represents a Device sampling frequency in samples/second.
type SampleRate int

// String returns a human-readable representation of sampling rate.
func (s SampleRate) String() {
	return fmt.Sprintf("%s samples/s", fntVal(float64(s)))
}

// Device represents a connected sampling device (e.g. USB oscilloscope),
type Device interface {
	// String returns a description of the device. It should be specific enough
	// to allow the user to identify the physical device that this value
	// represents.
	String() string

	// Channels returns a map of Channels indexed by their IDs. Channel can be used
	// to configure parameters related to a single capture source.
	Channels() map[ChanID]Channel

	// ReadData asks the device for a trace.
	// This interface assumes all channels on a single Device are sampled at the
	// same rate and return the same number of samples for every run.
	ReadData() (map[ChanID][]byte, time.Duration, error)

	// GetSampleRates returns a slice of sample rates available on this device.
	GetSampleRates() []SampleRate
}
