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
	r       scope.DataRecorder
	stop    chan struct{}
	chans   map[scope.ChanID]dataSrc
	chanIDs []scope.ChanID
}

func (dum) String() string { return "dummy device" }

func (d *dum) Channels() []scope.ChanID {
	return d.chanIDs
}

func (d *dum) Attach(rec scope.DataRecorder) {
	d.r = rec
}

func (d *dum) Start() {
	d.stop = make(chan struct{}, 1)
	ch := make(chan []scope.ChannelData)
	d.r.Reset(scope.Millisecond, ch)
	go func() {
		offset := rand.Intn(200)
		for {
			var dat []scope.ChannelData
			for _, s := range d.chanIDs {
				dat = append(dat, scope.ChannelData{
					ID:      s,
					Samples: d.chans[s].data(offset),
				})
			}
			select {
			case ch <- dat:
			case <-d.stop:
				close(ch)
				return
			}
		}
	}()
}

func (d *dum) Stop() {
	d.stop <- struct{}{}
}
