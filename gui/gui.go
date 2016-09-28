package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
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

func (p aggrPoint) add(x, y int) aggrPoint {
	p.x = x
	p.sumY += y
	p.sizeY++
	return p
}

func (p aggrPoint) toPoint() image.Point {
	return image.Point{p.x, p.sumY / p.sizeY}
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
	for i, y := range s {
		mapX := int(float64(start.X) + float64(i)/rangeX*float64(end.X-start.X))
		mapY := int(float64(end.Y) - float64(y-minY)/rangeY*float64(end.Y-start.Y))
		aggrPoints[mapX] = aggrPoints[mapX].add(mapX, mapY)
	}
	fmt.Println(aggrPoints)
	var points []image.Point
	for _, p := range aggrPoints {
		points = append(points, p.toPoint())
	}

	return points
}

// Represents the entire plotting area.
type Plot struct {
	*image.RGBA
}

// Fills the plot with a color.
func (plot Plot) Fill(col color.RGBA) {
	bounds := plot.Bounds()
	for i := bounds.Min.X; i < bounds.Max.X; i++ {
		for j := bounds.Min.Y; j < bounds.Max.Y; j++ {
			plot.Set(i, j, col)
		}
	}
}

// Draws a straight line from pixel p1 to p2.
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

// Draws samples in the image rectangle defined by
// starting (upper left) and ending (lower right) pixel.
func (plot Plot) DrawSamples(start, end image.Point, s []scope.Sample, col color.RGBA) {
	points := samplesToPoints(s, start, end)
	sort.Sort(xSorter(points))
	for i := 1; i < len(points); i++ {
		plot.DrawLine(points[i-1], points[i], col)
	}
}

// Draws samples from all the channels into one image.
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

type xSorter []image.Point

func (a xSorter) Len() int {
	return len(a)
}

func (a xSorter) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a xSorter) Less(i, j int) bool {
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
	dum, _ := dummy.Open("")
	data, stop, _ := dum.StartSampling()

	samples := (<-data).Samples
	fmt.Println(samples)
	stop()

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
	for _, id := range dum.Channels() {
		cols[id] = chanCols[next]
		next = (next + 1) % 4
	}
	plot.DrawAll(samples, cols)

	f, err := os.Create("draw.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	png.Encode(f, plot)
}
