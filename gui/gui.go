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
	"image/png"
	"os"

	"github.com/zagrodzki/goscope/scope"
)

const (
	divRows            = 8
	divCols            = 10
	defaultZero        = 0.5
	defaultVoltsPerDiv = 0.5
)

var colorWhite = color.RGBA{255, 255, 255, 255}
var colorBlack = color.RGBA{0, 0, 0, 255}

type aggrPoint struct {
	sumY  int
	sizeY int
}

func (p *aggrPoint) add(y int) {
	p.sumY += y
	p.sizeY++
}

func (p *aggrPoint) toPoint(x int) image.Point {
	return image.Point{x, p.sumY / p.sizeY}
}

// TracePos represents the position of zero and volts per div
type TracePos struct {
	// the position of Y=0 (0 <= Zero <= 1) given as
	// the fraction of the window height counting from the bottom
	Zero float64
	// volts per div
	PerDiv float64
}

func samplesToPoints(samples []scope.Sample, tracePos TracePos, start, end image.Point) []image.Point {
	if len(samples) == 0 {
		return nil
	}

	sampleMaxY := (1 - tracePos.Zero) * divRows * tracePos.PerDiv
	sampleMinY := -tracePos.Zero * divRows * tracePos.PerDiv
	sampleWidthX := float64(len(samples) - 1)
	sampleWidthY := sampleMaxY - sampleMinY

	pixelStartX := float64(start.X)
	pixelEndY := float64(end.Y - 1)
	pixelWidthX := float64(end.X - start.X - 1)
	pixelWidthY := float64(end.Y - start.Y - 1)
	ratioX := pixelWidthX / sampleWidthX
	ratioY := pixelWidthY / sampleWidthY

	points := make([]image.Point, end.X-start.X)
	lastAggr := aggrPoint{}
	lastX := start.X
	pi := 0
	for i, y := range samples {
		mapX := int(pixelStartX + float64(i)*ratioX)
		mapY := int(pixelEndY - float64(y-scope.Sample(sampleMinY))*ratioY)
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

func isInside(x, y int, start, end image.Point) bool {
	return x >= start.X && x <= end.X && y >= start.Y && y <= end.Y
}

// DrawLine draws a straight line from pixel p1 to p2.
// Only the line fragment inside the image rectangle defined by
// starting (upper left) and ending (lower right) pixel is drawn.
func (plot Plot) DrawLine(p1, p2 image.Point, start, end image.Point, col color.RGBA) {
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
			y := int(a*float64(i) + b)
			if isInside(i, y, start, end) {
				plot.SetRGBA(i, y, col)
			}
		}
	} else {
		// If the line is more vertical than horizontal,
		// for every pixel row between p1 and p2
		// we find and switch on the pixel closest to y=a*x+b
		for i := min(p1.Y, p2.Y); i <= max(p1.Y, p2.Y); i++ {
			x := int((float64(i) - b) / a)
			if isInside(x, i, start, end) {
				plot.SetRGBA(x, i, col)
			}
		}
	}
}

// DrawSamples draws samples in the image rectangle defined by
// starting (upper left) and ending (lower right) pixel.
func (plot Plot) DrawSamples(samples []scope.Sample, tracePos TracePos, start, end image.Point, col color.RGBA) {
	points := samplesToPoints(samples, tracePos, start, end)
	for i := 1; i < len(points); i++ {
		plot.DrawLine(points[i-1], points[i], start, end, col)
	}
}

// DrawAll draws samples from all the channels in the plot.
func (plot Plot) DrawAll(samples map[scope.ChanID][]scope.Sample, tracePos map[scope.ChanID]TracePos, cols map[scope.ChanID]color.RGBA) {
	plot.Fill(colorWhite)
	b := plot.Bounds()
	for id, v := range samples {
		pos, exists := tracePos[id]
		if !exists {
			pos = TracePos{defaultZero, defaultVoltsPerDiv}
		}
		col, exists := cols[id]
		if !exists {
			col = colorBlack
		}
		plot.DrawSamples(v, pos, b.Min, b.Max, col)
	}
}

// DrawFromDevice draws samples from the device in the plot.
func (plot Plot) DrawFromDevice(dev scope.Device, tracePos map[scope.ChanID]TracePos, cols map[scope.ChanID]color.RGBA) error {
	data, stop, err := dev.StartSampling()
	defer stop()
	if err != nil {
		return err
	}
	samples := (<-data).Samples
	plot.DrawAll(samples, tracePos, cols)
	return nil
}

// CreatePlot plots samples from the device.
func CreatePlot(dev scope.Device, width, height int, tracePos map[scope.ChanID]TracePos, cols map[scope.ChanID]color.RGBA) (Plot, error) {
	plot := Plot{image.NewRGBA(image.Rect(0, 0, width, height))}
	err := plot.DrawFromDevice(dev, tracePos, cols)
	return plot, err
}

// PlotToPng creates a plot of the samples from the device
// and saves it as PNG.
func PlotToPng(dev scope.Device, width, height int, tracePos map[scope.ChanID]TracePos, cols map[scope.ChanID]color.RGBA, outputFile string) error {
	plot, err := CreatePlot(dev, width, height, tracePos, cols)
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
