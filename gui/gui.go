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


type Plot struct {
    *image.RGBA
}

func (plot Plot) Fill(col color.RGBA) {
    bounds := plot.Bounds()
    for i := bounds.Min.X; i < bounds.Max.X; i++ {
        for j := bounds.Min.Y; j < bounds.Max.Y; j++ {
            plot.Set(i, j, col)
        }
    }
}

func Min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func Max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

func Abs(a int) int {
    if a < 0 {
        return -a
    }
    return a
}

type AggrPoint struct {
    x int
    sumY int
    sizeY int
}

func (p AggrPoint) add(x, y int) AggrPoint {
    p.x = x
    p.sumY += y
    p.sizeY++
    return p
}

func (p AggrPoint) toPoint() image.Point {
    return image.Point{p.x, p.sumY / p.sizeY}
}

func SamplesToPoints(s []scope.Sample, start, end image.Point) []image.Point {
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

    rangeX := float64(len(s)-1)
    rangeY := float64(maxY - minY)
    aggrPoints := make(map[int]AggrPoint)
    for i, y := range s {
        mapX := int(float64(start.X) + float64(i) / rangeX * float64(end.X - start.X))
        mapY := int(float64(end.Y) - float64(y - minY) / rangeY * float64(end.Y - start.Y))
        aggrPoints[mapX] = aggrPoints[mapX].add(mapX, mapY)
    }
    fmt.Println(aggrPoints)
    var points []image.Point
    for _, p := range aggrPoints {
        points = append(points, p.toPoint())
    }

    return points
}

func (plot Plot) DrawLine(p1, p2 image.Point, col color.RGBA) {
    if p1.X == p2.X {
        for i := Min(p1.Y, p2.Y); i <= Max(p1.Y, p2.Y); i++ {
            plot.Set(p1.X, i, col)
        }
        return
    }
    a := float64(p1.Y - p2.Y) / float64(p1.X - p2.X)
    b := float64(p1.Y) - float64(p1.X) * a

    if Abs(p1.X - p2.X) >= Abs(p1.Y - p2.Y) {
        for i := Min(p1.X, p2.X); i <= Max(p1.X, p2.X); i++ {
            plot.Set(i, int(a * float64(i) + b), col)
        }
    } else {
        for i := Min(p1.Y, p2.Y); i <= Max(p1.Y, p2.Y); i++ {
            plot.Set(int((float64(i) - b) / a), i, col)
        }
    }
}

type XSorter []image.Point

func (a XSorter) Len() int           { return len(a) }
func (a XSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a XSorter) Less(i, j int) bool { return a[i].X < a[j].X }


func (plot Plot) DrawSamples(start, end image.Point, s []scope.Sample, col color.RGBA) {
    points := SamplesToPoints(s, start, end)
    sort.Sort(XSorter(points))
    for i := 1; i < len(points); i++ {
        plot.DrawLine(points[i-1], points[i], col)
    }
}

func (plot Plot) DrawAll(samples map[scope.ChanID][]scope.Sample, cols map[scope.ChanID]color.RGBA) {
    b := plot.Bounds()
    x1 := b.Min.X + 10
    x2 := b.Max.X - 10
    y1 := b.Min.Y + 10
    y2 := b.Min.Y + 10 + int((b.Max.Y - b.Min.Y - 10 * (len(samples) + 1)) / len(samples))
    step := y2 - b.Min.Y
    for id, v := range samples {
        plot.DrawSamples(image.Point{x1, y1}, image.Point{x2, y2}, v, cols[id])
        y1 = y1 + step
        y2 = y2 + step
    }
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
