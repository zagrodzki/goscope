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
    img *image.RGBA
}

func (plot Plot) Fill(col color.RGBA) {
    for i := plot.img.Bounds().Min.X; i < plot.img.Bounds().Max.X; i++ {
        for j := plot.img.Bounds().Min.Y; j < plot.img.Bounds().Max.Y; j++ {
            plot.img.Set(i, j, col)
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

func LinePoints(p1, p2 image.Point) []image.Point {
    points := make([]image.Point, 0)
    if p1.X == p2.X {
        for i := Min(p1.Y, p2.Y); i <= Max(p1.Y, p2.Y); i++ {
            points = append(points, image.Point{p1.X, i})
        }
        return points
    }
    a := (float64(p1.Y) - float64(p2.Y)) / (float64(p1.X) - float64(p2.X))
    b := float64(p1.Y) - float64(p1.X) * a

    if Abs(p1.X - p2.X) >= Abs(p1.Y - p2.Y) {
        for i := Min(p1.X, p2.X); i <= Max(p1.X, p2.X); i++ {
            points = append(points, image.Point{i, int(a * float64(i) + b)})
        }
    } else {
        for i := Min(p1.Y, p2.Y); i <= Max(p1.Y, p2.Y); i++ {
            points = append(points, image.Point{int((float64(i) - b) / a), i})
        }
    }
    return points
}



func SamplesToPoints(s []scope.Sample, start, end image.Point) []image.Point {
    min_y := s[0]
    max_y := s[0]
    for _, y := range s {
        if min_y > y {
            min_y = y
        }
        if max_y < y {
            max_y = y
        }
    }
    range_x := float64(len(s)-1)
    range_y := float64(max_y - min_y)
    points := make([]image.Point, 0)
    sums := make(map[int]int)
    sizes := make(map[int]int)
    for i, y := range s {
        mapx := int(float64(start.X) + (float64(i) / range_x) * float64(end.X - start.X))
        mapy := int(float64(end.Y) - ((float64(y) - float64(min_y)) / range_y) * float64(end.Y - start.Y))
        sums[mapx] = sums[mapx]+mapy
        sizes[mapx] = sizes[mapx]+1
    }
    var keys []int
    for k := range sums {
        keys = append(keys, k)
    }
    sort.Ints(keys)
    for _, i := range keys {
        points = append(points, image.Point{i, int(float64(sums[i])/float64(sizes[i]))})
    }

    return points
}

func (plot Plot) DrawPoints(points []image.Point, col color.RGBA) {
    for _, p := range points {
        plot.img.Set(p.X, p.Y, col)
    }
}

func (plot Plot) DrawSamples(start, end image.Point, s []scope.Sample, col color.RGBA) {
    points := SamplesToPoints(s, start, end)
    plot.DrawPoints(points, col)
    for i := 1; i < len(points); i++ {
        plot.DrawPoints(LinePoints(points[i-1], points[i]), col)
    }
}

func (plot Plot) DrawAll(samples map[scope.ChanID][]scope.Sample, col color.RGBA) {
    b := plot.img.Bounds()
    x1 := b.Min.X + 10
    x2 := b.Max.X - 10
    y1 := b.Min.Y + 10
    y2 := b.Min.Y + 10 + int((b.Max.Y - b.Min.Y - 10 * (len(samples) + 1)) / len(samples))
    step := y2 - b.Min.Y
    for _, v := range samples {
        plot.DrawSamples(image.Point{x1, y1}, image.Point{x2, y2}, v, col)
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

    plot := Plot{img: image.NewRGBA(image.Rect(0, 0, 800, 600))}
    colWhite := color.RGBA{255, 255, 255, 255}
    colRed := color.RGBA{255, 0, 0, 255}
    plot.Fill(colWhite)
    plot.DrawAll(samples, colRed)

    f, err := os.Create("draw.png")
    if err != nil {
        panic(err)
    }
    defer f.Close()
    png.Encode(f, plot.img)
}
