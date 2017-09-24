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
	"flag"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/zagrodzki/goscope/compat"
	"github.com/zagrodzki/goscope/scope"
)

const (
	DivRows            = 8
	DivCols            = 10
	defaultZero        = 0.5
	defaultVoltsPerDiv = 0.5
)

var colorWhite = color.RGBA{255, 255, 255, 255}
var colorBlack = color.RGBA{0, 0, 0, 255}

var interpType = flag.String("interpolation", "sinczeropad", "interpolation type: one of linear, step, sinc, sinczeropad")

func interpolator() Interpolator {
	switch *interpType {
	case "linear":
		return LinearInterpolator
	case "step":
		return StepInterpolator
	case "sinc":
		return SincInterpolator
	case "sinczeropad":
		return SincZeroPadInterpolator
	}
	log.Fatalf("Invalid value %q for flag \"interpolation\", want one of: linear, step, sinc, sinczeropad", *interpType)
	return nil
}

type aggrPoint struct {
	sumY  float64
	sizeY int
}

func (p *aggrPoint) add(y float64) {
	p.sumY += y
	p.sizeY++
}

func (p *aggrPoint) toPoint(x int) image.Point {
	return image.Point{x, round(p.sumY / float64(p.sizeY))}
}

func samplesToPoints(samples []scope.Voltage, traceParams scope.TraceParams, rect image.Rectangle) []image.Point {
	if len(samples) == 0 {
		return nil
	}

	sampleMaxY := (1 - traceParams.Zero) * DivRows * traceParams.PerDiv
	sampleMinY := -traceParams.Zero * DivRows * traceParams.PerDiv
	sampleWidthX := float64(len(samples) - 1)
	sampleWidthY := sampleMaxY - sampleMinY

	pixelStartX := float64(rect.Min.X)
	pixelEndY := float64(rect.Max.Y - 1)
	pixelWidthX := float64(rect.Dx() - 1)
	pixelWidthY := float64(rect.Dy() - 1)
	ratioX := pixelWidthX / sampleWidthX
	ratioY := pixelWidthY / sampleWidthY

	points := make([]image.Point, rect.Dx())
	lastAggr := aggrPoint{}
	lastX := rect.Min.X
	pi := 0
	for i, y := range samples {
		mapX := round(pixelStartX + float64(i)*ratioX)
		mapY := pixelEndY - float64(y-scope.Voltage(sampleMinY))*ratioY
		if lastX != mapX {
			points[pi] = lastAggr.toPoint(lastX)
			pi++
			lastX = mapX
			lastAggr = aggrPoint{}
		}
		lastAggr.add(mapY)
	}
	points[pi] = lastAggr.toPoint(lastX)
	pi++

	return points[:pi]
}

// Plot represents the entire plotting area.
type Plot struct {
	*image.RGBA
	interp Interpolator
}

// NewPlot returns a new Plot of specified size.
func NewPlot(p image.Point) Plot {
	return Plot{
		image.NewRGBA(image.Rect(0, 0, p.X, p.Y)),
		interpolator(),
	}
}

var (
	bgCache *image.RGBA
	bgColor color.RGBA
)

func background(r image.Rectangle, col color.RGBA) *image.RGBA {
	img := image.NewRGBA(r)
	pix := img.Pix
	for i := 0; i < len(pix); i = i + 4 {
		pix[i] = col.R
		pix[i+1] = col.G
		pix[i+2] = col.B
		pix[i+3] = col.A
	}
	return img
}

// Fill fills the plot with a background image of the same size.
func (plot Plot) Fill(col color.RGBA) {
	if bgCache == nil || bgCache.Bounds() != plot.Bounds() || bgColor != col {
		bgCache = background(plot.Bounds(), col)
		bgColor = col
	}
	copy(plot.Pix, bgCache.Pix)
}

func isInside(x, y int, rect image.Rectangle) bool {
	return x >= rect.Min.X && x <= rect.Max.X && y >= rect.Min.Y && y <= rect.Max.Y
}

// DrawLine draws a straight line from pixel p1 to p2.
// Only the line fragment inside the image rectangle defined by
// starting (upper left) and ending (lower right) pixel is drawn.
func (plot Plot) DrawLine(p1, p2 image.Point, rect image.Rectangle, col color.RGBA) {
	if p1.X == p2.X { // vertical line
		for i := min(p1.Y, p2.Y); i <= max(p1.Y, p2.Y); i++ {
			plot.SetRGBA(p1.X, i, col)
		}
		return
	}

	// Calculating the parameters of the equation
	// of the straight line (in the form y=a*x+b)
	// passing through p1 and p2.

	// slope of the line
	a := float64(p1.Y-p2.Y) / float64(p1.X-p2.X)
	// intercept of the line
	b := float64(p1.Y) - float64(p1.X)*a

	// To avoid visual "gaps" between the pixels we switch on,
	// we draw the line in one of two ways.
	if abs(p1.X-p2.X) >= abs(p1.Y-p2.Y) {
		// If the line is more horizontal than vertical,
		// for every pixel column between p1 and p2
		// we find and switch on the pixel closest to y=a*x+b
		for i := min(p1.X, p2.X); i <= max(p1.X, p2.X); i++ {
			y := round(a*float64(i) + b)
			if isInside(i, y, rect) {
				plot.SetRGBA(i, y, col)
			}
		}
	} else {
		// If the line is more vertical than horizontal,
		// for every pixel row between p1 and p2
		// we find and switch on the pixel closest to y=a*x+b
		for i := min(p1.Y, p2.Y); i <= max(p1.Y, p2.Y); i++ {
			x := round((float64(i) - b) / a)
			if isInside(x, i, rect) {
				plot.SetRGBA(x, i, col)
			}
		}
	}
}

// DrawSamples draws samples in the image rectangle defined by
// starting (upper left) and ending (lower right) pixel.
func (plot Plot) DrawSamples(samples []scope.Voltage, traceParams scope.TraceParams, rect image.Rectangle, col color.RGBA) error {
	if len(samples) < rect.Dx() {
		interpSamples, err := plot.interp(samples, rect.Dx())
		if err != nil {
			return err
		}
		samples = interpSamples
	}
	points := samplesToPoints(samples, traceParams, rect)
	for i := 1; i < len(points); i++ {
		plot.DrawLine(points[i-1], points[i], rect, col)
	}
	return nil
}

// DrawAll draws samples from all the channels in the plot.
func (plot Plot) DrawAll(data []scope.ChannelData, traceParams map[scope.ChanID]scope.TraceParams, cols map[scope.ChanID]color.RGBA) error {
	b := plot.Bounds()
	for _, chanData := range data {
		id, v := chanData.ID, chanData.Samples
		params, exists := traceParams[id]
		if !exists {
			params = scope.TraceParams{defaultZero, defaultVoltsPerDiv}
		}
		col, exists := cols[id]
		if !exists {
			col = colorBlack
		}
		if err := plot.DrawSamples(v, params, b, col); err != nil {
			return err
		}
	}
	return nil
}

// DrawFromDevice draws samples from the device in the plot.
func (plot Plot) DrawFromDevice(dev scope.Device, traceParams map[scope.ChanID]scope.TraceParams, cols map[scope.ChanID]color.RGBA) error {
	rec := &compat.Recorder{TB: scope.Millisecond}
	dev.Attach(rec)
	dev.Start()
	defer dev.Stop()
	samples := (<-rec.Data).Channels
	return plot.DrawAll(samples, traceParams, cols)
}

// CreatePlot plots samples from the device.
func CreatePlot(dev scope.Device, width, height int, traceParams map[scope.ChanID]scope.TraceParams, cols map[scope.ChanID]color.RGBA) (Plot, error) {
	plot := Plot{
		RGBA:   image.NewRGBA(image.Rect(0, 0, width, height)),
		interp: SincInterpolator,
	}
	err := plot.DrawFromDevice(dev, traceParams, cols)
	return plot, err
}

// PlotToPng creates a plot of the samples from the device
// and saves it as PNG.
func PlotToPng(dev scope.Device, width, height int, traceParams map[scope.ChanID]scope.TraceParams, cols map[scope.ChanID]color.RGBA, outputFile string) error {
	plot, err := CreatePlot(dev, width, height, traceParams, cols)
	if err != nil {
		return err
	}

	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()
	png.Encode(f, plot)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func round(a float64) int {
	return int(a + 0.5)
}
