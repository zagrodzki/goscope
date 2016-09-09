package dummy

import (
	"time"

	"bitbucket.org/zagrodzki/goscope/scope"
)

type dum struct{}

func (dum) String() string                     { return "dummy device" }
func (dum) GetSampleRate() scope.SampleRate    { return 1000 }
func (dum) GetSampleRates() []scope.SampleRate { return []scope.SampleRate{1000} }
func (dum) SetSampleRate() error               { return nil }

func (dum) Channels() map[scope.ChanID]scope.Channel {
	return map[scope.ChanID]scope.Channel{
		"zero":     zeroChan{},
		"sin":      sinChan{},
		"square":   squareChan{},
		"triangle": triangleChan{},
	}
}

func (dum) ReadData() (map[scope.ChanID][]scope.Sample, time.Duration, error) {
	return map[scope.ChanID][]scope.Sample{
		"zero":     zeroChan{}.data(),
		"sin":      sinChan{}.data(),
		"square":   squareChan{}.data(),
		"triangle": triangleChan{}.data(),
	}, time.Millisecond, nil
}
