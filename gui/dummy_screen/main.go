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

package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"sync"
	"time"

	"github.com/zagrodzki/goscope/dummy"
	"github.com/zagrodzki/goscope/gui"
	"github.com/zagrodzki/goscope/scope"
	"github.com/zagrodzki/goscope/triggers"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/time/rate"
)

var (
	triggerSource    = flag.String("trigger_source", "", "Name of the channel to use as a trigger source")
	triggerThresh    = flag.Float64("trigger_threshold", 0, "Trigger threshold")
	triggerEdge      = flag.String("trigger_edge", "rising", "Trigger edge, rising or falling")
	useChan          = flag.String("channel", "sin", "one of the channels of dummy device: zero,random,sin,triangle,square")
	timeBase         = flag.Duration("timebase", time.Second, "timebase of the displayed waveform")
	perDiv           = flag.Float64("v_per_div", 2, "volts per div")
	screenWidth      = flag.Int("width", 800, "UI width, in pixels")
	screenHeight     = flag.Int("height", 600, "UI height, in pixels")
	refreshRateLimit = flag.Float64("refresh_rate", 25, "maximum refresh rate, in frames per second")
)

type waveform struct {
	tb    scope.Duration
	inter scope.Duration
	tp    map[scope.ChanID]scope.TraceParams

	mu      sync.Mutex
	plot    gui.Plot
	bufPlot gui.Plot
}

func (w *waveform) TimeBase() scope.Duration {
	return w.tb
}

var allColors = []color.RGBA{
	color.RGBA{255, 0, 0, 255},
	color.RGBA{0, 200, 0, 255},
	color.RGBA{0, 0, 255, 255},
	color.RGBA{255, 0, 255, 255},
	color.RGBA{255, 255, 0, 255},
}

func (w *waveform) swapPlot() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.plot, w.bufPlot = w.bufPlot, w.plot
}

func (w *waveform) keepReading(dataCh <-chan []scope.ChannelData) {
	var buf []scope.ChannelData
	var tbCount = int(w.tb / w.inter)
	chColor := make(map[scope.ChanID]color.RGBA)
	for data := range dataCh {
		if len(data) == 0 {
			continue
		}
		if buf == nil {
			buf = make([]scope.ChannelData, len(data))
			for i, d := range data {
				buf[i].ID = d.ID
				buf[i].Samples = make([]scope.Voltage, 0, 2*tbCount)
				chColor[d.ID] = allColors[i]
			}
		}
		for i, d := range data {
			buf[i].Samples = append(buf[i].Samples, d.Samples...)
		}
		if len(buf[0].Samples) >= tbCount {
			for i := range data {
				buf[i].Samples = buf[i].Samples[:tbCount]
			}

			// full timebase, draw and go to beginning
			w.bufPlot.DrawAll(buf, w.tp, chColor)
			w.swapPlot()
			// truncate the buffers
			for i := range buf {
				buf[i].Samples = buf[i].Samples[:0]
			}
		}
	}
}

func (w *waveform) Reset(inter scope.Duration, d <-chan []scope.ChannelData) {
	w.inter = inter
	go w.keepReading(d)
}

func (w *waveform) Error(error) {}

func (w *waveform) SetTimeBase(d scope.Duration) {
	w.tb = d
}

func (w *waveform) SetChannel(ch scope.ChanID, p scope.TraceParams) {
	if w.tp == nil {
		w.tp = make(map[scope.ChanID]scope.TraceParams)
	}
	w.tp[ch] = p
}

func (w *waveform) Render() *image.RGBA {
	w.mu.Lock()
	defer w.mu.Unlock()
	ret := image.NewRGBA(w.plot.RGBA.Rect)
	copy(ret.Pix, w.plot.RGBA.Pix)
	return ret
}

func main() {
	flag.Parse()

	var edge triggers.RisingEdge
	switch *triggerEdge {
	case "rising":
		edge = triggers.Rising
	case "falling":
		edge = triggers.Falling
	default:
		log.Fatalf("Unknown value %q for flag trigger_edge, expected rising or falling", *triggerEdge)
	}

	dev, _ := dummy.Open(*useChan)

	screenSize := image.Point{*screenWidth, *screenHeight}
	wf := &waveform{
		plot:    gui.NewPlot(screenSize),
		bufPlot: gui.NewPlot(screenSize),
	}
	wf.SetTimeBase(scope.DurationFromNano(*timeBase))

	for _, id := range dev.Channels() {
		wf.SetChannel(id, scope.TraceParams{Zero: 0.5, PerDiv: *perDiv})
	}

	tr := triggers.New(wf)
	tr.Source(scope.ChanID(*triggerSource))
	tr.Edge(edge)
	tr.Level(scope.Voltage(*triggerThresh))

	dev.Attach(tr)
	dev.Start()
	defer dev.Stop()

	driver.Main(func(s screen.Screen) {
		w, err := s.NewWindow(&screen.NewWindowOptions{Width: screenSize.X, Height: screenSize.Y})
		if err != nil {
			log.Fatalf("NewWindow: %v", err)
		}
		defer w.Release()
		stop := make(chan struct{})
		go processEvents(w, stop)

		b, err := s.NewBuffer(screenSize)
		if err != nil {
			log.Fatalf("NewBuffer(): %v", err)
		}
		defer b.Release()
		limiter := rate.NewLimiter(rate.Limit(*refreshRateLimit), 1)
		sometimes := rate.NewLimiter(0.2, 1)
		for {
			select {
			case <-stop:
				return
			default:
			}
			limiter.Wait(context.Background())
			t := time.Now()
			trace := wf.Render()
			if trace == nil {
				continue
			}
			copy(b.RGBA().Pix, trace.Pix)
			w.Upload(image.Point{0, 0}, b, b.Bounds())
			w.Publish()
			if sometimes.Allow() {
				d := time.Since(t)
				fmt.Printf("Rendering 1 frame took %v (%.2ffps)\n", d, float64(time.Second)/float64(d))
			}
		}
	})
}
