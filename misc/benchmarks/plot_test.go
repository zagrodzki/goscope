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

package benchmark

import (
	"testing"

	"github.com/zagrodzki/goscope/dummy"
	"github.com/zagrodzki/goscope/gui"
	"github.com/zagrodzki/goscope/scope"
)

func BenchmarkGuiFull(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dev, err := dummy.Open("")
		if err != nil {
			b.Fatalf("Cannot open the device: %v", err)
		}
		err = gui.PlotToPng(dev, make(map[scope.ChanID]gui.ChannelYRange), "draw.png")
		if err != nil {
			b.Fatalf("Cannot plot to file: %v", err)
		}
	}
}

func BenchmarkGuiOnlyPlot(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dev, err := dummy.Open("")
		if err != nil {
			b.Fatalf("Cannot open the device: %v", err)
		}
		_, err = gui.CreatePlot(dev, make(map[scope.ChanID]gui.ChannelYRange))
		if err != nil {
			b.Fatalf("Cannot create plot: %v", err)
		}
	}
}
