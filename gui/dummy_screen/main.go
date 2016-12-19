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
	"os"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"github.com/zagrodzki/goscope/dummy"
	"github.com/zagrodzki/goscope/gui"
	"github.com/zagrodzki/goscope/scope"
	"github.com/zagrodzki/goscope/triggers"
	"github.com/zagrodzki/goscope/usb"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/time/rate"
)

var (
	device           = flag.String("device", "", "Device to use, autodetect if empty")
	list             = flag.Bool("list", false, "If set, only list available devices")
	triggerSource    = flag.String("trigger_source", "", "Name of the channel to use as a trigger source")
	triggerThresh    = flag.Float64("trigger_threshold", 0, "Trigger threshold")
	triggerEdge      = flag.String("trigger_edge", "rising", "Trigger edge, rising or falling")
	triggerMode      = flag.String("trigger_mode", "auto", "Trigger mode, auto, single or normal")
	useChan          = flag.String("channel", "sin", "one of the channels of dummy device: zero,random,sin,triangle,square")
	timeBase         = flag.Duration("timebase", time.Second, "timebase of the displayed waveform")
	perDiv           = flag.Float64("v_per_div", 2, "volts per div")
	screenWidth      = flag.Int("width", 800, "UI width, in pixels")
	screenHeight     = flag.Int("height", 600, "UI height, in pixels")
	refreshRateLimit = flag.Float64("refresh_rate", 25, "maximum refresh rate, in frames per second. 0 = no limit")
	cpuprofile       = flag.String("cpuprofile", "", "File to which the program should write it's CPU profile (performance stats)")
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

func (w *waveform) Render(ret *image.RGBA) {
	w.mu.Lock()
	defer w.mu.Unlock()
	copy(ret.Pix, w.plot.RGBA.Pix)
}

type system struct {
	name      string
	enumerate func() map[string]string
	open      func(string) (scope.Device, error)
}

var (
	systems = []system{
		{
			name:      "dummy",
			enumerate: dummy.Enumerate,
			open:      dummy.Open,
		},
		{
			name:      "usb",
			enumerate: usb.Enumerate,
			open:      usb.Open,
		},
	}
	systemsByName = make(map[string]int)
)

func parseTriggerModeFlag() triggers.Mode {
	switch *triggerMode {
	case "auto":
		return triggers.ModeAuto
	case "normal":
		return triggers.ModeNormal
	case "single":
		return triggers.ModeSingle
	}
	log.Fatalf("Unknown value %q for flag trigger_mode, expected auto, normal or single", *triggerMode)
	return triggers.ModeNone
}

func main() {
	flag.Parse()

	mode := parseTriggerModeFlag()

	var all []string
	for idx, sys := range systems {
		systemsByName[sys.name] = idx
		for id := range sys.enumerate() {
			all = append(all, fmt.Sprintf("%s:%s", sys.name, id))
		}
	}

	if len(all) == 0 {
		log.Fatalf("Did not find any supported devices")
	}
	if *list {
		fmt.Println("Devices found:")
		for _, d := range all {
			fmt.Println(d)
		}
		return
	}
	id := all[0]
	if *device != "" {
		for _, d := range all {
			if d == *device {
				id = d
				break
			}
		}
		if id != *device {
			log.Fatalf("Device %s not detected on the list. Available devices: %v", *device, all)
		}
	} else if len(all) > 1 {
		log.Printf("Multiple devices found: %v", all)
		log.Printf("Using the first device (%s)", id)
	}

	parts := strings.SplitN(id, ":", 2)
	s := systemsByName[parts[0]]
	osc, err := systems[s].open(parts[1])

	if err != nil {
		log.Fatalf("Open: %+v", err)
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	screenSize := image.Point{*screenWidth, *screenHeight}
	wf := &waveform{
		plot:    gui.NewPlot(screenSize),
		bufPlot: gui.NewPlot(screenSize),
	}
	wf.SetTimeBase(scope.DurationFromNano(*timeBase))

	for _, id := range osc.Channels() {
		wf.SetChannel(id, scope.TraceParams{Zero: 0.5, PerDiv: *perDiv})
	}

	tr := triggers.New(wf)
	tr.Source(scope.ChanID(*triggerSource))
	tr.Level(scope.Voltage(*triggerThresh))
	tr.Mode(mode)
	// For now, the names of params are hardcoded here, but in the future
	// names might change between devices and it's not very practical.
	// The intention is to have params initialized to defaults and then changed
	// only through the UI or by specifying the parameter name and value
	// on the commandline. But because there is no UI yet, it's not really feasible.
	for _, p := range tr.TriggerParams() {
		switch pn := p.Name(); pn {
		case "Trigger edge":
			if err := p.Set(*triggerEdge); err != nil {
				log.Fatalf("TriggerParams[%q].Set(%q): %v", pn, *triggerEdge, err)
			}
		}
	}

	osc.Attach(tr)
	osc.Start()
	defer osc.Stop()

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

		var limiter *rate.Limiter
		if *refreshRateLimit > 0 {
			limiter = rate.NewLimiter(rate.Limit(*refreshRateLimit), 1)
		}
		sometimes := rate.NewLimiter(0.2, 1)
		for {
			select {
			case <-stop:
				return
			default:
			}
			if limiter != nil {
				limiter.Wait(context.Background())
			}
			t := time.Now()
			wf.Render(b.RGBA())
			w.Upload(image.Point{0, 0}, b, b.Bounds())
			w.Publish()
			if sometimes.Allow() {
				d := time.Since(t)
				fmt.Printf("Rendering 1 frame took %v (%.2ffps)\n", d, float64(time.Second)/float64(d))
			}
		}
	})
}
