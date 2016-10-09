//  Copyright 2016 The goscope Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package dummy

import (
	"github.com/zagrodzki/goscope/scope"
)

type dum struct {
	enabled map[scope.ChanID]bool
	chans   []scope.ChanID
}

func (dum) String() string                     { return "dummy device" }
func (dum) GetSampleRate() scope.SampleRate    { return 1000 }
func (dum) GetSampleRates() []scope.SampleRate { return []scope.SampleRate{1000} }
func (dum) SetSampleRate() error               { return nil }

func (d dum) Channels() []scope.ChanID {
	return d.chans
}

func (dum) Channel(ch scope.ChanID) scope.Channel {
	switch ch {
	case "zero":
		return zeroChan{}
	case "sin":
		return sinChan{}
	case "square":
		return squareChan{}
	case "triangle":
		return triangleChan{}
	case "random":
		return randomChan{}
	}
	return nil
}

func (d dum) StartSampling() (<-chan scope.Data, func(), error) {
	stop := make(chan struct{}, 1)
	data := make(chan scope.Data)
	go func() {
		for {
			dat := scope.Data{
				Samples:  make(map[scope.ChanID][]scope.Sample),
				Num:      numSamples,
				Interval: scope.Millisecond,
			}
			samples := map[scope.ChanID][]scope.Sample{
				"zero":     zeroChan{}.data(),
				"sin":      sinChan{}.data(),
				"square":   squareChan{}.data(),
				"triangle": triangleChan{}.data(),
				"random":   randomChan{}.data(),
			}
			for _, ch := range d.chans {
				if s, ok := samples[ch]; ok {
					dat.Samples[ch] = s
				}
			}
			select {
			case <-stop:
				return
			case data <- dat:
			}
		}
	}()
	return data, func() { stop <- struct{}{} }, nil
}
