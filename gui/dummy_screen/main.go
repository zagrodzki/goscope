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
	"image"
	"image/color"
	"log"

	"github.com/zagrodzki/goscope/dummy"
	"github.com/zagrodzki/goscope/gui"
	"github.com/zagrodzki/goscope/scope"
	"github.com/zagrodzki/goscope/triggers"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/time/rate"
)

const (
	screenWidth      = 1200
	screenHeight     = 600
	refreshRateLimit = 25
)

var (
	hasTrigger = flag.Bool("enable_trigger", true, "When true, a trigger is configured on the device")
	useChan    = flag.String("channel", "sin", "one of the channels of dummy device: zero,random,sin,triangle,square")
)

func main() {
	flag.Parse()
	dev, _ := dummy.Open(*useChan)
	data, stop, err := dev.StartSampling()
	if err != nil {
		log.Fatalf("StartSampling(): %v", err)
	}
	defer stop()

	zas := make(map[scope.ChanID]gui.TracePos)
	cols := make(map[scope.ChanID]color.RGBA)
	for _, id := range dev.Channels() {
		zas[id] = gui.TracePos{Zero: 0.5, PerDiv: 2}
		cols[id] = color.RGBA{0, 0, 255, 255}
	}

	driver.Main(func(s screen.Screen) {
		w, err := s.NewWindow(&screen.NewWindowOptions{Width: screenWidth, Height: screenHeight})
		if err != nil {
			log.Fatalf("NewWindow: %v", err)
		}
		defer w.Release()
		stop := make(chan struct{})
		go processEvents(w, stop)

		b, err := s.NewBuffer(image.Point{screenWidth, screenHeight})
		if err != nil {
			log.Fatalf("NewBuffer(): %v", err)
		}
		defer b.Release()
		limiter := rate.NewLimiter(rate.Limit(refreshRateLimit), 1)
		var in <-chan scope.Data
		if *hasTrigger {
			newCh := make(chan scope.Data, 10)
			tr := triggers.New(data, newCh)
			in = newCh
			tr.TimeBase(scope.Millisecond)
			tr.Source(scope.ChanID(*useChan))
			tr.Edge(triggers.Falling)
			tr.Level(-0.9)
		} else {
			in = data
		}
		for {
			select {
			case <-stop:
				return
			default:
			}
			limiter.Wait(context.Background())
			get := <-in
			p := gui.Plot{b.RGBA()}
			for i := 0; i < len(p.Pix); i++ {
				p.Pix[i] = 255
			}
			p.DrawAll(get.Samples, zas, cols)
			w.Upload(image.Point{0, 0}, b, b.Bounds())
			w.Publish()
		}
	})
}
