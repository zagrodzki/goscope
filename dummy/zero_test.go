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
	"testing"

	"github.com/zagrodzki/goscope/scope"
)

func TestZero(t *testing.T) {
	ch := zeroChan{}
	data := ch.data(0)
	for _, tc := range []struct {
		idx  int
		want scope.Sample
	}{
		// zero is always 0
		{0, 0},
		{100, 0},
		{999, 0},
	} {
		if got := data[tc.idx]; got != tc.want {
			t.Errorf("zero.data()[%d]: got %v, want %v", tc.idx, got, tc.want)
		}
	}
}
