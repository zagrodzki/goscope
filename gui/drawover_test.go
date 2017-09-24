//  Copyright 2017 The goscope Authors
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
	"image/draw"
	"testing"
)

func bgAndPlot() (*image.RGBA, *image.RGBA) {
	size := image.Point{800, 600}
	bg := NewPlot(size)
	bg.Fill(ColorWhite)
	plot := NewPlot(size)
	plot.DrawLine(image.Point{0, 0}, size, plot.Bounds(), ColorBlack)
	plot.DrawLine(image.Point{0, size.Y}, image.Point{size.X, 0}, plot.Bounds(), ColorRed)
	plot.DrawLine(image.Point{0, size.Y / 2}, image.Point{size.X, size.Y / 2}, plot.Bounds(), ColorGreen)
	plot.DrawLine(image.Point{size.X / 2, 0}, image.Point{size.X / 2, size.Y}, plot.Bounds(), ColorBlue)
	return bg.RGBA, plot.RGBA
}

func TestDrawOver(t *testing.T) {
	// Assuming that image/draw.Draw is the canonical implementation and
	// is always correct.
	bg, plot := bgAndPlot()
	if plot.Opaque() {
		t.Fatal("plot is completely opaque, expected partial translucent")
	}
	want := image.NewRGBA(bg.Bounds())
	draw.Draw(want, want.Bounds(), bg, image.Point{0, 0}, draw.Over)
	draw.Draw(want, want.Bounds(), plot, image.Point{0, 0}, draw.Over)
	got := image.NewRGBA(bg.Bounds())
	DrawOver(got, bg)
	DrawOver(got, plot)
	for i := 0; i < len(got.Pix); i += 4 {
		for j := i; j < i+4; j++ {
			if got.Pix[j] != want.Pix[j] {
				t.Fatalf("pixel %d: got %v, want %v", i, got.Pix[i:i+3], want.Pix[i:i+3])
			}
		}
	}
}

func BenchmarkDrawOver(b *testing.B) {
	bg, plot := bgAndPlot()
	b.Run("draw package dense", func(b *testing.B) {
		out := image.NewRGBA(bg.Bounds())
		for i := 0; i < b.N; i++ {
			draw.Draw(out, out.Bounds(), bg, image.Point{0, 0}, draw.Over)
			draw.Draw(out, out.Bounds(), plot, image.Point{0, 0}, draw.Over)
		}
	})
	b.Run("simplified dense", func(b *testing.B) {
		out := image.NewRGBA(bg.Bounds())
		for i := 0; i < b.N; i++ {
			DrawOver(out, bg)
			DrawOver(out, plot)
		}
	})
	b.Run("draw package sparse", func(b *testing.B) {
		out := image.NewRGBA(bg.Bounds())
		for i := 0; i < b.N; i++ {
			draw.Draw(out, out.Bounds(), plot, image.Point{0, 0}, draw.Over)
		}
	})
	b.Run("simplified sparse", func(b *testing.B) {
		out := image.NewRGBA(bg.Bounds())
		for i := 0; i < b.N; i++ {
			DrawOver(out, plot)
		}
	})
}
