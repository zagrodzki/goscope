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

type dataSrc interface {
	data(int) []scope.Voltage
}

type dum struct {
	chans   map[scope.ChanID]dataSrc
	chanIDs []scope.ChanID
}

func (dum) String() string { return "dummy device" }

func (d dum) Channels() []scope.ChanID {
	return d.chanIDs
}

func (d dum) StartSampling() (<-chan scope.Data, func(), error) {
	stop := make(chan struct{}, 1)
	data := make(chan scope.Data)
	go func() {
		for {
			dat := scope.Data{
				Num:      numSamples,
				Interval: scope.Millisecond,
			}
			offset := rand.Intn(200)
			for _, s := range d.chanIDs {
				dat.Channels = append(dat.Channels, scope.ChannelData{
					ID:      s,
					Samples: d.chans[s].data(offset),
				})
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
