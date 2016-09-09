package dummy

import "bitbucket.org/zagrodzki/goscope/scope"

type triangleChan struct{}

func (triangleChan) ID() scope.ChanID                   { return "triangle" }
func (triangleChan) GetVoltRange() scope.VoltRange      { return 1 }
func (triangleChan) GetVoltRanges() []scope.VoltRange   { return []scope.VoltRange{1} }
func (triangleChan) SetVoltRange(scope.VoltRange) error { return nil }
func (triangleChan) data() []scope.Sample {
	ret := make([]scope.Sample, numSamples)
	for i := 0; i < numSamples; i++ {
		if i%40 < 20 {
			ret[i] = scope.Sample(float64(i%20-10) / 10)
		} else {
			ret[i] = scope.Sample(float64(30-i%40) / 10)
		}
	}
	return ret
}
