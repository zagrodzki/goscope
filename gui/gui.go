package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"sort"

	"github.com/zagrodzki/goscope/dummy"
	"github.com/zagrodzki/goscope/scope"
)

type aggrPoint struct {
	x     int
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

func samplesToPoints(s []scope.Sample, start, end image.Point) []image.Point {
	if len(s) == 0 {
		return nil
	}
	minY := s[0]
	maxY := s[0]
	for _, y := range s {
		if minY > y {
			minY = y
		}
		if maxY < y {
			maxY = y
		}
	}

	rangeX := float64(len(s) - 1)
	rangeY := float64(maxY - minY)
	aggrPoints := make(map[int]aggrPoint)
	startX := float64(start.X)
	endY := float64(end.Y)
	pixelRangeX := float64(end.X - start.X)
	pixelRangeY := float64(end.Y - start.Y)
	for i, y := range s {
		mapX := int(startX + float64(i)/rangeX*pixelRangeX)
		mapY := int(endY - float64(y-minY)/rangeY*pixelRangeY)
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

// DrawLine draws a straight line from pixel p1 to p2.
func (plot Plot) DrawLine(p1, p2 image.Point, col color.RGBA) {
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
			plot.Set(i, int(a*float64(i)+b), col)
		}
	} else {
		// If the line is more vertical than horizontal,
		// for every pixel row between p1 and p2
		// we find and switch on the pixel closest to y=a*x+b
		for i := min(p1.Y, p2.Y); i <= max(p1.Y, p2.Y); i++ {
			plot.Set(int((float64(i)-b)/a), i, col)
		}
	}
}

// DrawSamples draws samples in the image rectangle defined by
// starting (upper left) and ending (lower right) pixel.
func (plot Plot) DrawSamples(start, end image.Point, s []scope.Sample, col color.RGBA) {
	points := samplesToPoints(s, start, end)
	sort.Sort(pointsByX(points))
	for i := 1; i < len(points); i++ {
		plot.DrawLine(points[i-1], points[i], col)
	}
}

// DrawAll draws samples from all the channels into one image.
func (plot Plot) DrawAll(samples map[scope.ChanID][]scope.Sample, cols map[scope.ChanID]color.RGBA) {
	b := plot.Bounds()
	x1 := b.Min.X + 10
	x2 := b.Max.X - 10
	y1 := b.Min.Y + 10
	y2 := b.Min.Y + 10 + int((b.Max.Y-b.Min.Y-10*(len(samples)+1))/len(samples))
	step := y2 - b.Min.Y
	for id, v := range samples {
		plot.DrawSamples(image.Point{x1, y1}, image.Point{x2, y2}, v, cols[id])
		y1 = y1 + step
		y2 = y2 + step
	}
}

func plot(dev scope.Device, outputFile string) error {
	data, stop, err := dev.StartSampling()
	defer stop()
	if err != nil {
		return err
	}
	samples := (<-data).Samples

	plot := Plot{image.NewRGBA(image.Rect(0, 0, 800, 600))}
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
		cols[id] = chanCols[next]
		next = (next + 1) % 4
	}
	plot.DrawAll(samples, cols)

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

func main() {
	dev, err := dummy.Open("")
	if err != nil {
		log.Fatalf("Cannot open the device", err)
	}
	if err := plot(dev, "draw.png"); err != nil {
		log.Fatalf("Cannot plot samples", err)
	}
}
