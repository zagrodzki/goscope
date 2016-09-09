package dummy

import (
	"math"

	"bitbucket.org/zagrodzki/goscope/scope"
)

type sinChan struct{}

func (sinChan) ID() scope.ChanID                   { return "sin" }
func (sinChan) GetVoltRange() scope.VoltRange      { return 1 }
func (sinChan) GetVoltRanges() []scope.VoltRange   { return []scope.VoltRange{1} }
func (sinChan) SetVoltRange(scope.VoltRange) error { return nil }
func (sinChan) data() []scope.Sample {
	ret := make([]scope.Sample, numSamples)
	for i := 0; i < numSamples; i++ {
		ret[i] = scope.Sample(math.Sin(float64(i) / 5))
	}
	return ret
}
