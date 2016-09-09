// Package scope defines an abstract interface for a digital oscilloscope or other
// similar capture device.
package scope

import (
	"fmt"
	"time"
)

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

    // GetSampleRate returns the currently configured sample rate.
    GetSampleRate() SampleRate

	// GetSampleRates returns a slice of sample rates available on this device.
	GetSampleRates() []SampleRate
}
