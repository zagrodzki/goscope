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
	err = gui.PlotToPng(dev, make(map[scope.ChanID]gui.ZeroAndScale), "plot1.png")
	if err != nil {
		log.Fatalf("Cannot plot to file: %v", err)
	}
	zas := map[scope.ChanID]gui.ZeroAndScale{
		"square":   gui.ZeroAndScale{0.1, 5},
		"triangle": gui.ZeroAndScale{0.8, 2},
	}
	err = gui.PlotToPng(dev, zas, "plot2.png")
	if err != nil {
		log.Fatalf("Cannot plot to file: %v", err)
	}
}
