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

package main

import (
	"image/color"
	"log"

	"github.com/zagrodzki/goscope/dummy"
	"github.com/zagrodzki/goscope/gui"
	"github.com/zagrodzki/goscope/scope"
)

func main() {
	dev, err := dummy.Open("")
	if err != nil {
		log.Fatalf("Cannot open the device: %v", err)
	}
	err = gui.PlotToPng(dev, 800, 600,
		make(map[scope.ChanID]gui.ZeroAndScale),
		make(map[scope.ChanID]color.RGBA),
		"plot1.png")
	if err != nil {
		log.Fatalf("Cannot plot to file: %v", err)
	}
	zas := map[scope.ChanID]gui.ZeroAndScale{
		"square":   gui.ZeroAndScale{0.1, 5},
		"triangle": gui.ZeroAndScale{0.8, 2},
	}
	cols := map[scope.ChanID]color.RGBA{
		"random":   color.RGBA{255, 0, 0, 255},
		"sin":      color.RGBA{255, 0, 255, 255},
		"square":   color.RGBA{0, 255, 0, 255},
		"triangle": color.RGBA{0, 0, 255, 255},
	}
	err = gui.PlotToPng(dev, 800, 600, zas, cols, "plot2.png")
	if err != nil {
		log.Fatalf("Cannot plot to file: %v", err)
	}
}
