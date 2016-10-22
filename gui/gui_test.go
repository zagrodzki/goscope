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
// 1) has the same bounds as the true plot
// 2) is entirely contained in the true plot
// 3) contains at least one pixel in every column
// 4) contains at least minPointCount pixels
func evaluatePlot(truePlot, testPlot image.Image, minPointCount int) (bool, string) {
	if truePlot.Bounds() != testPlot.Bounds() {
		return false, fmt.Sprintf("plot bounds: got %v, expected %v", testPlot.Bounds(), truePlot.Bounds())
	}
	b := truePlot.Bounds()
	pointCount := 0
	for x := b.Min.X; x < b.Max.X; x++ {
		col := false
		for y := b.Min.Y; y < b.Max.Y; y++ {
			testOn := isOn(testPlot, x, y)
			col = col || testOn
			if testOn {
				pointCount++
			}
			if testOn && !isOn(truePlot, x, y) {
				return false, "test plot is not contained in true plot"
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

// testPlot evaluates a plot generated from the samples against a true plot stored in a file.
// minPointCount is the minimum number of pixels of the tested plot.
// minSimilarity is the minimum similarity of the plots.
func testPlot(t *testing.T, plotFile string, samples []scope.Sample, minPointCount int) {
	file, err := os.Open(plotFile)
	if err != nil {
		t.Fatalf("Cannot open file: %v", err)
	}
	img, err := png.Decode(file)
	if err != nil {
		t.Fatalf("Cannot decode file: %v", err)
	}

	plot := Plot{image.NewRGBA(image.Rect(0, 0, 800, 600))}
	plot.Fill(colorWhite)
	plotBounds := plot.Bounds()
	plot.DrawSamples(samples, TracePos{0.5, 0.25}, plotBounds.Min, plotBounds.Max, colorBlack)
	eval, msg := evaluatePlot(img, plot, minPointCount)
	if !eval {
		t.Errorf(msg)
	}
}

func TestSin(t *testing.T) {
	numSamples := 1000
	interval := 4 * math.Pi / float64(numSamples-1)
	samples := make([]scope.Sample, numSamples)
	for i := 0; i < numSamples; i++ {
		samples[i] = scope.Sample(math.Sin(float64(i) * interval))
	}
	testPlot(t, "sin-gp.png", samples, 2000)
}

func TestZero(t *testing.T) {
	numSamples := 1000
	samples := make([]scope.Sample, numSamples)
	for i := 0; i < numSamples; i++ {
		samples[i] = 0
	}
	testPlot(t, "zero-gp.png", samples, 800)
}

func TestSquare(t *testing.T) {
	numSamples := 1000
	samples := make([]scope.Sample, numSamples)
	for i := 0; i < numSamples/4; i++ {
		samples[i] = 1
		samples[i+numSamples/4] = -1
		samples[i+numSamples/2] = 1
		samples[i+3*numSamples/4] = -1
	}
	testPlot(t, "square-gp.png", samples, 2000)
}

func TestTriangle(t *testing.T) {
	numSamples := 999
	interval := 2.0 / float64(numSamples/3-1)
	samples := make([]scope.Sample, numSamples)
	for i := 0; i < numSamples/3; i++ {
		offset := float64(i) * interval
		samples[i] = scope.Sample(-1.0 + offset)
		samples[i+numSamples/3] = scope.Sample(1.0 - offset)
		samples[i+2*numSamples/3] = scope.Sample(-1.0 + offset)
	}
	testPlot(t, "triangle-gp.png", samples, 1000)
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
