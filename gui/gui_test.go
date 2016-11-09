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
func evaluatePlot(refPlot, testPlot image.Image, minPointCount int) error {
	if got, want := testPlot.Bounds(), refPlot.Bounds(); got != want {
		return fmt.Errorf("plot bounds: got %v, want %v", got, want)
	}

	b := refPlot.Bounds()
	pointCount := 0
	for x := b.Min.X; x < b.Max.X; x++ {
		col := false
		for y := b.Min.Y; y < b.Max.Y; y++ {
			testOn := isOn(testPlot, x, y)
			if testOn {
				pointCount++
				col = true
			}
			if testOn && !isOn(refPlot, x, y) {
				return fmt.Errorf("point (%v, %v) of the test plot is not marked on the reference plot", x, y)
			}
		}
		if !col {
			return fmt.Errorf("image column %v does not contain any point", x)
		}
	}
	if got, want := pointCount, minPointCount; got < want {
		return fmt.Errorf("too few plot points: got %v, want at least %v", got, want)
	}
	return nil
}

func TestPlot(t *testing.T) {
	for _, tc := range []struct {
		desc          string
		numSamples    int
		gen           func(int) scope.Sample
		interp        Interpolator
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
		{
			desc:       "sin interpolated",
			numSamples: 20,
			gen: func(i int) scope.Sample {
				return scope.Sample(-0.7 * math.Sin(16*float64(i)*math.Pi/20))
			},
			interp:        SincInterpolator,
			minPointCount: 2000,
			refPlotFile:   "sin2-gp.png",
		},
		{
			desc:       "sin sum interpolated",
			numSamples: 15,
			gen: func(i int) scope.Sample {
				return scope.Sample(0.5*math.Sin(8*float64(i)*math.Pi/15) + 0.5*math.Sin(12*float64(i)*math.Pi/15))
			},
			interp:        SincInterpolator,
			minPointCount: 2000,
			refPlotFile:   "sin-sum-gp.png",
		},
		{
			desc:       "square interpolated",
			numSamples: 4,
			gen: func(i int) scope.Sample {
				return scope.Sample(-2*(i%2) + 1)
			},
			interp:        ConstInterpolator,
			minPointCount: 2000,
			refPlotFile:   "square-short-gp.png",
		},
		{
			desc:       "lines interpolated",
			numSamples: 7,
			gen: func(i int) scope.Sample {
				switch i {
				case 1:
					return 0.5
				case 2:
					return 1
				case 5:
					return -1
				case 6:
					return 1
				default:
					return 0
				}

			},
			interp:        LinearInterpolator,
			minPointCount: 1000,
			refPlotFile:   "lines-gp.png",
		},
	} {
		samples := make([]scope.Sample, tc.numSamples)
		for i := 0; i < tc.numSamples; i++ {
			samples[i] = tc.gen(i)
		}
		file, err := os.Open(tc.refPlotFile)
		if err != nil {
			t.Errorf("Cannot open file %v: %v", tc.refPlotFile, err)
			continue
		}
		refPlot, err := png.Decode(file)
		if err != nil {
			t.Errorf("Cannot decode file %v: %v", tc.refPlotFile, err)
			continue
		}

		testPlot := Plot{image.NewRGBA(image.Rect(0, 0, 800, 600))}
		testPlot.Fill(colorWhite)
		b := testPlot.Bounds()
		testPlot.DrawSamples(samples, TraceParams{0.5, 0.25, tc.interp}, b, colorBlack)
		err = evaluatePlot(refPlot, testPlot, tc.minPointCount)
		if err != nil {
			t.Errorf("error in evaluating plot %v against %v: %v", tc.desc, tc.refPlotFile, err)
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
		make(map[scope.ChanID]TraceParams),
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
	traceParams := map[scope.ChanID]TraceParams{
		"square":   TraceParams{0.1, 5, SincInterpolator},
		"triangle": TraceParams{0.8, 2, SincInterpolator},
	}
	cols := map[scope.ChanID]color.RGBA{
		"random":   color.RGBA{255, 0, 0, 255},
		"sin":      color.RGBA{255, 0, 255, 255},
		"square":   color.RGBA{0, 255, 0, 255},
		"triangle": color.RGBA{0, 0, 255, 255},
	}
	err = PlotToPng(dev, 800, 600, traceParams, cols, filepath.Join(dir, "plot.png"))
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
			make(map[scope.ChanID]TraceParams),
			make(map[scope.ChanID]color.RGBA))
		if err != nil {
			b.Fatalf("Cannot create plot: %v", err)
		}
	}
}
