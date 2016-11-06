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
	"flag"
	"log"

	"github.com/zagrodzki/goscope/dummy"
	"github.com/zagrodzki/goscope/gui"
)

var (
	fileName = flag.String("file", "draw.png", "output file name")
	width    = flag.Int("width", 800, "PNG width")
	height   = flag.Int("height", 600, "PNG width")
	tracePos = posFlag("tpos", "zero and volts per div, format: \"chanID:zero,perDiv\"")
	cols     = colFlag("col", "color, format: \"chanID:R,G,B\"")
)

func main() {
	flag.Parse()
	dev, err := dummy.Open("")
	if err != nil {
		log.Fatalf("Cannot open the device: %v", err)
	}

	err = gui.PlotToPng(dev, *width, *height, nil, *tracePos, *cols, *fileName)
	if err != nil {
		log.Fatalf("Cannot plot to file: %v", err)
	}
}
