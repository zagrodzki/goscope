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
	"sort"

	"github.com/zagrodzki/goscope/scope"
)

type aggrPoint struct {
	sumY  int
	sizeY int
}

func (p aggrPoint) add(y int) aggrPoint {
	p.sumY += y
	p.sizeY++
	return p
}

func (p aggrPoint) toPoint(x int) image.Point {
	return image.Point{x, p.sumY / p.sizeY}
}

// ChannelYRange represents a y-range of samples to be plotted
type ZeroAndScale struct {
	// the position of Y=0 (0 <= Zero <= 1) given as
	// the fraction of the window height counting from the top
	Zero float64
	// scale of the plot in sample units per pixel
	Scale float64
}

func samplesToPoints(samples []scope.Sample, zeroAndScale ZeroAndScale, start, end image.Point) []image.Point {
	if len(samples) == 0 {
		return nil
	}

	sampleMaxY := zeroAndScale.Zero * zeroAndScale.Scale
	sampleMinY := (zeroAndScale.Zero - 1) * zeroAndScale.Scale
	sampleWidthX := float64(len(samples) - 1)
	sampleWidthY := sampleMaxY - sampleMinY

	pixelStartX := float64(start.X)
	pixelEndY := float64(end.Y)
	pixelWidthX := float64(end.X - start.X)
	pixelWidthY := float64(end.Y - start.Y)

	aggrPoints := make(map[int]aggrPoint)
	for i, y := range samples {
		mapX := int(pixelStartX + float64(i)/sampleWidthX*pixelWidthX)
		mapY := int(pixelEndY - float64(y-scope.Sample(sampleMinY))/sampleWidthY*pixelWidthY)
		aggrPoints[mapX] = aggrPoints[mapX].add(mapY)
	}
	var points []image.Point
	for x, p := range aggrPoints {
		points = append(points, p.toPoint(x))
	}

	return points
}

// Plot represents the entire plotting area.
type Plot struct {
	*image.RGBA
}

// Fill fills the plot with a color.
func (plot Plot) Fill(col color.RGBA) {
	bounds := plot.Bounds()
	for i := bounds.Min.X; i < bounds.Max.X; i++ {
		for j := bounds.Min.Y; j < bounds.Max.Y; j++ {
			plot.Set(i, j, col)
		}
	}
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
			plot.Set(p1.X, i, col)
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
				plot.Set(i, y, col)
			}
		}
	} else {
		// If the line is more vertical than horizontal,
		// for every pixel row between p1 and p2
		// we find and switch on the pixel closest to y=a*x+b
		for i := min(p1.Y, p2.Y); i <= max(p1.Y, p2.Y); i++ {
			x := int((float64(i) - b) / a)
			if isInside(x, i, start, end) {
				plot.Set(x, i, col)
			}
		}
	}
}

// DrawSamples draws samples in the image rectangle defined by
// starting (upper left) and ending (lower right) pixel.
func (plot Plot) DrawSamples(samples []scope.Sample, zeroAndScale ZeroAndScale, start, end image.Point, col color.RGBA) {
	points := samplesToPoints(samples, zeroAndScale, start, end)
	sort.Sort(pointsByX(points))
	for i := 1; i < len(points); i++ {
		plot.DrawLine(points[i-1], points[i], start, end, col)
	}
}

// DrawAll draws samples from all the channels into one image.
func (plot Plot) DrawAll(samples map[scope.ChanID][]scope.Sample, zas map[scope.ChanID]ZeroAndScale, cols map[scope.ChanID]color.RGBA) {
	b := plot.Bounds()
	x1 := b.Min.X + 10
	x2 := b.Max.X - 10
	y1 := b.Min.Y + 10
	y2 := b.Min.Y + 10 + int((b.Max.Y-b.Min.Y-10*(len(samples)+1))/len(samples))
	step := y2 - b.Min.Y
	for id, v := range samples {
		plot.DrawSamples(v, zas[id], image.Point{x1, y1}, image.Point{x2, y2}, cols[id])
		y1 = y1 + step
		y2 = y2 + step
	}
}

// CreatePlot plots samples from the device.
func CreatePlot(dev scope.Device, zas map[scope.ChanID]ZeroAndScale) (Plot, error) {
	plot := Plot{image.NewRGBA(image.Rect(0, 0, 800, 600))}

	data, stop, err := dev.StartSampling()
	defer stop()
	if err != nil {
		return plot, err
	}
	samples := (<-data).Samples

	colWhite := color.RGBA{255, 255, 255, 255}
	colRed := color.RGBA{255, 0, 0, 255}
	colGreen := color.RGBA{0, 255, 0, 255}
	colBlue := color.RGBA{0, 0, 255, 255}
	colBlack := color.RGBA{0, 0, 0, 255}
	chanCols := [4]color.RGBA{colRed, colGreen, colBlue, colBlack}

	plot.Fill(colWhite)

	cols := make(map[scope.ChanID]color.RGBA)
	next := 0
	for _, id := range dev.Channels() {
		if _, exists := zas[id]; !exists {
			zas[id] = ZeroAndScale{0.5, 2}
		}
		cols[id] = chanCols[next]
		next = (next + 1) % 4
	}
	plot.DrawAll(samples, zas, cols)
	return plot, nil
}

// PlotToPng creates a plot of the samples from the device
// and saves it as PNG.
func PlotToPng(dev scope.Device, zas map[scope.ChanID]ZeroAndScale, outputFile string) error {
	plot, err := CreatePlot(dev, zas)
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

type pointsByX []image.Point

func (a pointsByX) Len() int {
	return len(a)
}

func (a pointsByX) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a pointsByX) Less(i, j int) bool {
	return a[i].X < a[j].X
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
