package dummy

import "bitbucket.org/zagrodzki/goscope/scope"

type squareChan struct{}

func (squareChan) ID() scope.ChanID                   { return "square" }
func (squareChan) GetVoltRange() scope.VoltRange      { return 1 }
func (squareChan) GetVoltRanges() []scope.VoltRange   { return []scope.VoltRange{1} }
func (squareChan) SetVoltRange(scope.VoltRange) error { return nil }
func (squareChan) data() []scope.Sample {
	ret := make([]scope.Sample, numSamples)
	for i := 0; i < numSamples; i++ {
		ret[i] = scope.Sample(1 - 2*((i/20)%2))
	}
	return ret
}
