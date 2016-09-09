package dummy

import "bitbucket.org/zagrodzki/goscope/scope"

type zeroChan struct{}

func (zeroChan) ID() scope.ChanID                   { return "sin" }
func (zeroChan) GetVoltRange() scope.VoltRange      { return 1 }
func (zeroChan) GetVoltRanges() []scope.VoltRange   { return []scope.VoltRange{1} }
func (zeroChan) SetVoltRange(scope.VoltRange) error { return nil }
func (zeroChan) data() []scope.Sample {
	return make([]scope.Sample, numSamples)
}
