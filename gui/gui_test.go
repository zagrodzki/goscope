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
	"image"
	"image/color"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/zagrodzki/goscope/dummy"
	"github.com/zagrodzki/goscope/scope"
)

// TODO: more tests for specific functions
// TODO: tests comparing the resulting plot image to the perfect image

func TestPlotToPng(t *testing.T) {
	dev, err := dummy.Open("")
	if err != nil {
		t.Fatalf("Cannot open the device: %v", err)
	}
	dir, err := ioutil.TempDir("", "TestPlotToPng")
	if err != nil {
		t.Fatalf("Cannot create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)
	err = PlotToPng(dev, 800, 600,
		make(map[scope.ChanID]ZeroAndScale),
		make(map[scope.ChanID]color.RGBA),
		filepath.Join(dir, "plot.png"))
	if err != nil {
		t.Fatalf("Cannot plot to file: %v", err)
	}
}

func TestPlotToPngWithCustomParameters(t *testing.T) {
	dev, err := dummy.Open("")
	if err != nil {
		t.Fatalf("Cannot open the device: %v", err)
	}
	dir, err := ioutil.TempDir("", "TestPlotToPngWithCustomScales")
	if err != nil {
		t.Fatalf("Cannot create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)
	zas := map[scope.ChanID]ZeroAndScale{
		"square":   ZeroAndScale{0.1, 5},
		"triangle": ZeroAndScale{0.8, 2},
	}
	cols := map[scope.ChanID]color.RGBA{
		"random":   color.RGBA{255, 0, 0, 255},
		"sin":      color.RGBA{255, 0, 255, 255},
		"square":   color.RGBA{0, 255, 0, 255},
		"triangle": color.RGBA{0, 0, 255, 255},
	}
	err = PlotToPng(dev, 800, 600, zas, cols, filepath.Join(dir, "plot.png"))
	if err != nil {
		t.Fatalf("Cannot plot to file: %v", err)
	}
}

func BenchmarkCreatePlot(b *testing.B) {
	dev, err := dummy.Open("")
	if err != nil {
		b.Fatalf("Cannot open the device: %v", err)
	}
	plot := Plot{image.NewRGBA(image.Rect(0, 0, 800, 600))}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = plot.DrawFromDevice(dev,
			make(map[scope.ChanID]ZeroAndScale),
			make(map[scope.ChanID]color.RGBA))
		if err != nil {
			b.Fatalf("Cannot create plot: %v", err)
		}
	}
}
