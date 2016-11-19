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
	"reflect"
	"testing"

	"github.com/zagrodzki/goscope/scope"
)

func TestDummy(t *testing.T) {
	dev, _ := Open("zero")
	data, stop, _ := dev.StartSampling()
	defer stop()
	d := <-data
	want := map[scope.ChanID]int{
		"zero": 300,
	}
	wantChans := make(map[scope.ChanID]bool)
	for k := range want {
		wantChans[k] = true
	}
	got := make(map[scope.ChanID]int)
	for _, ch := range d.Channels {
		got[ch.ID] = len(ch.Samples)
	}
	gotChans := make(map[scope.ChanID]bool)
	for k := range got {
		gotChans[k] = true
	}

	if !reflect.DeepEqual(gotChans, wantChans) {
		t.Errorf("got data for channels: %v, want %v", gotChans, wantChans)
	}
	for k, v := range want {
		if got[k] < v {
			t.Errorf("samples for channel %v: got %d, want at least %d", k, got[k], v)
		}
	}
}
