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

package gui

import (
	"testing"

	"github.com/zagrodzki/goscope/dummy"
	"github.com/zagrodzki/goscope/scope"
)

func TestPlotToPng(t *testing.T) {
	dev, err := dummy.Open("")
	if err != nil {
		t.Fatalf("Cannot open the device: %v", err)
	}
	err = PlotToPng(dev, make(map[scope.ChanID]ChannelYRange), "plot1.png")
	if err != nil {
		t.Fatalf("Cannot plot to file: %v", err)
	}
}

func TestPlotToPngWithRanges(t *testing.T) {
	dev, err := dummy.Open("")
	if err != nil {
		t.Fatalf("Cannot open the device: %v", err)
	}
	ranges := map[scope.ChanID]ChannelYRange{
		"square":   ChannelYRange{0, 1},
		"triangle": ChannelYRange{-2, 2},
	}
	err = PlotToPng(dev, ranges, "plot2.png")
	if err != nil {
		t.Fatalf("Cannot plot to file: %v", err)
	}
}
