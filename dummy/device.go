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
