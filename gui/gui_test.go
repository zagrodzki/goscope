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
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/zagrodzki/goscope/dummy"
	"github.com/zagrodzki/goscope/scope"
)

// TODO: more tests for specific functions

// isOn returns true if pixel x,y is of a different color than white
func isOn(img image.Image, x, y int) bool {
	return img.At(x, y) != colorWhite
}

// evaluatePlot checks whether the tested plot:
// 1) has the same bounds as the reference plot
// 2) is entirely contained in the reference plot
// 3) contains at least one pixel in every column
// 4) contains at least minPointCount pixels
func evaluatePlot(refPlot, testPlot image.Image, minPointCount int) (bool, string) {
	if refPlot.Bounds() != testPlot.Bounds() {
		return false, fmt.Sprintf("plot bounds: got %v, expected %v", testPlot.Bounds(), refPlot.Bounds())
	}
	b := refPlot.Bounds()
	pointCount := 0
	for x := b.Min.X; x < b.Max.X; x++ {
		col := false
		for y := b.Min.Y; y < b.Max.Y; y++ {
			testOn := isOn(testPlot, x, y)
			col = col || testOn
			if testOn {
				pointCount++
			}
			if testOn && !isOn(refPlot, x, y) {
				return false, "test plot is not contained in reference plot"
			}
		}
		if !col {
			return false, fmt.Sprintf("image column %v does not contain any point", x)
		}
	}
	if pointCount < minPointCount {
		return false, fmt.Sprintf("too few plot points: got %v, expected at least %v", pointCount, minPointCount)
	}
	return true, ""
}

func TestPlot(t *testing.T) {
	for _, tc := range []struct {
		desc          string
		numSamples    int
		gen           func(int) scope.Sample
		minPointCount int
		refPlotFile   string
	}{
		{
			desc:       "sin",
			numSamples: 1000,
			gen: func(i int) scope.Sample {
				return scope.Sample(math.Sin(float64(i) * 4 * math.Pi / 999))
			},
			minPointCount: 2000,
			refPlotFile:   "sin-gp.png",
		},
		{
			desc:       "zero",
			numSamples: 1000,
			gen: func(i int) scope.Sample {
				return 0
			},
			minPointCount: 800,
			refPlotFile:   "zero-gp.png",
		},
		{
			desc:       "square",
			numSamples: 1000,
			gen: func(i int) scope.Sample {
				return scope.Sample(-2*(i/250%2) + 1)
			},
			minPointCount: 2000,
			refPlotFile:   "square-gp.png",
		},
		{
			desc:       "triangle",
			numSamples: 999,
			gen: func(i int) scope.Sample {
				sign := 2*(i/333%2) - 1
				return scope.Sample(float64(sign) * (1.0 - float64(i%333)*2.0/332.0))
			},
			minPointCount: 1000,
			refPlotFile:   "triangle-gp.png",
		},
	} {
		samples := make([]scope.Sample, tc.numSamples)
		for i := 0; i < tc.numSamples; i++ {
			samples[i] = tc.gen(i)
		}
		file, err := os.Open(tc.refPlotFile)
		if err != nil {
			t.Fatalf("Cannot open file: %v", err)
		}
		refPlot, err := png.Decode(file)
		if err != nil {
			t.Fatalf("Cannot decode file: %v", err)
		}

		testPlot := Plot{image.NewRGBA(image.Rect(0, 0, 800, 600))}
		testPlot.Fill(colorWhite)
		b := testPlot.Bounds()
		testPlot.DrawSamples(samples, TracePos{0.5, 0.25}, b.Min, b.Max, colorBlack)
		eval, msg := evaluatePlot(refPlot, testPlot, tc.minPointCount)
		if !eval {
			t.Errorf(fmt.Sprintf("error in evaluating plot %v: %v", tc.desc, msg))
		}
	}
}

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
		make(map[scope.ChanID]TracePos),
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
	tracePos := map[scope.ChanID]TracePos{
		"square":   TracePos{0.1, 5},
		"triangle": TracePos{0.8, 2},
	}
	cols := map[scope.ChanID]color.RGBA{
		"random":   color.RGBA{255, 0, 0, 255},
		"sin":      color.RGBA{255, 0, 255, 255},
		"square":   color.RGBA{0, 255, 0, 255},
		"triangle": color.RGBA{0, 0, 255, 255},
	}
	err = PlotToPng(dev, 800, 600, tracePos, cols, filepath.Join(dir, "plot.png"))
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
			make(map[scope.ChanID]TracePos),
			make(map[scope.ChanID]color.RGBA))
		if err != nil {
			b.Fatalf("Cannot create plot: %v", err)
		}
	}
}
