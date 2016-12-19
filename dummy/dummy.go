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
	"fmt"
	"log"
	"strings"

	"github.com/zagrodzki/goscope/scope"
	"github.com/zagrodzki/goscope/triggers"
)

const numSamples = 1000

// Enumerate returns the one and only dummy device
func Enumerate() map[string]string {
	log.Printf("Found: a dummy device")
	return map[string]string{
		"": "a dummy capture device",
	}
}

// Open opens the dummy device
func Open(ch string) (scope.Device, error) {
	if ch == "" {
		ch = "sin,triangle,square,random"
	}
	chNames := strings.Split(ch, ",")
	if got, want := len(chNames), 4; got > want {
		return nil, fmt.Errorf("device can have at most %d channels, got %d (%s)", got, want, ch)
	}
	d := &dum{
		chans: map[scope.ChanID]dataSrc{
			"zero":     zeroChan{},
			"sin":      sinChan{},
			"square":   squareChan{},
			"triangle": triangleChan{},
			"random":   &randomChan{},
		},
	}
	var chs []scope.ChanID
	for _, c := range chNames {
		cID := scope.ChanID(c)
		if _, ok := d.chans[cID]; !ok {
			return nil, fmt.Errorf("device does not have a channel named %s", c)
		}
		chs = append(chs, scope.ChanID(c))
	}
	d.chanIDs = chs
	return triggers.New(d), nil
}
