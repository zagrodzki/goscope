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

func TestRandom(t *testing.T) {
	origRand := randDiff
	defer func() { randDiff = origRand }()
	ch := &randomChan{}
	randDiff = func() scope.Voltage { return origRand() - ch.last }
	var sum scope.Voltage
	data := ch.data(0)
	for _, s := range data {
		sum += s
	}
	sum /= scope.Voltage(len(data))
	if sum > 0.1 || sum < -0.1 {
		t.Errorf("sum(normal distribution samples): expected 0+-0.1, got %v", sum)
	}
}
