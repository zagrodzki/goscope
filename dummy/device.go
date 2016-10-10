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
	"math/rand"

	"github.com/zagrodzki/goscope/scope"
)

type dum struct {
	chans    map[scope.ChanID]scope.Channel
	chanIDs  []scope.ChanID
	enabled  map[scope.ChanID]bool
	samplers map[scope.ChanID]func(int) []scope.Sample
}

func (dum) String() string                     { return "dummy device" }
func (dum) GetSampleRate() scope.SampleRate    { return 1000 }
func (dum) GetSampleRates() []scope.SampleRate { return []scope.SampleRate{1000} }
func (dum) SetSampleRate() error               { return nil }

func (d dum) Channels() []scope.ChanID {
	return d.chanIDs
}

func newChan(ch scope.ChanID) (scope.Channel, func(int) []scope.Sample) {
	switch ch {
	case "zero":
		return zeroChan{}, zeroChan{}.data
	case "sin":
		return sinChan{}, sinChan{}.data
	case "square":
		return squareChan{}, squareChan{}.data
	case "triangle":
		return triangleChan{}, triangleChan{}.data
	case "random":
		r := &randomChan{}
		return r, r.data
	}
	return nil, nil
}

func (d dum) Channel(ch scope.ChanID) scope.Channel {
	return d.chans[ch]
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
			offset := rand.Intn(200)
			for ch, s := range d.samplers {
				dat.Samples[ch] = s(offset)
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
