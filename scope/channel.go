package scope

import "fmt"

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

    // GetVoltRange returns the currently configured measurement range.
    GetVoltRange() VoltRange

	// GetVoltRanges returns a slice with available ranges that can be passed to SetVoltRange.
	GetVoltRanges() []VoltRange

	// SetVoltRange adjusts the sensitivity
	SetVoltRange(VoltRange) error
}
