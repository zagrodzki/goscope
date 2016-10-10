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
	want := map[scope.ChanID]bool{
		"zero": true,
	}
	got := make(map[scope.ChanID]bool)
	for ch := range d.Samples {
		got[ch] = true
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("sampling data: got channels %v, want %v", got, want)
	}
	if got, want := len(d.Samples["zero"]), 300; got < want {
		t.Errorf("sampling data: got %d samples for channel zero, want at least %d", got, want)
	}
}
