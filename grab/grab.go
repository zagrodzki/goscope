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
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/zagrodzki/goscope/dummy"
	"github.com/zagrodzki/goscope/scope"
	"github.com/zagrodzki/goscope/usb"
)

var (
	dev      = flag.String("device", "", "Device to use, autodetect if empty")
	list     = flag.Bool("list", false, "If set, only list available devices")
	chID     = flag.String("chan", "", "name of the channel to use. If not specified, use the first channel")
	period   = flag.Duration("period", 0, "how long period of samples to collect, run forever if set to 0")
	showHist = flag.Bool("histogram", false, "If true, output histogram of samples, otherwise only the mode")
)

func must(e error) {
	if e != nil {
		log.Fatalf(e.Error())
	}
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

type orderedHist struct {
	s map[scope.Voltage]int
	k []scope.Voltage
}

func (o *orderedHist) Len() int {
	return len(o.k)
}
func (o *orderedHist) Swap(i, j int) {
	o.k[i], o.k[j] = o.k[j], o.k[i]
}
func (o *orderedHist) Less(i, j int) bool {
	// sorting in reverse order
	return o.s[o.k[i]] > o.s[o.k[j]]
}
func (o *orderedHist) sort() {
	if len(o.k) != len(o.s) {
		o.k = make([]scope.Voltage, len(o.s))
		for s := range o.s {
			o.k = append(o.k, s)
		}
	}
	sort.Sort(o)
}

func main() {
	flag.Parse()
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
	if *dev != "" {
		for _, d := range all {
			if d == *dev {
				id = d
				break
			}
		}
		if id != *dev {
			log.Fatalf("Device %s not detected on the list. Available devices: %v", *dev, all)
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
	fmt.Println(osc)
	channels := osc.Channels()
	ch := channels[0]
	if *chID != "" {
		for _, c := range channels {
			if c == scope.ChanID(*chID) {
				ch = c
			}
		}
		if ch != scope.ChanID(*chID) {
			log.Fatalf("Device %s does not have a channel %q. Available channels: %v", id, *chID, channels)
		}
	}
	data, stop, err := osc.StartSampling()
	if err != nil {
		log.Fatalf("ReadData: %+v", err)
	}
	defer stop()
	i := int(scope.DurationFromNano(*period) / 1e6)
	log.Printf("Reading %d samples", i)
	for s := range data {
		hist := &orderedHist{
			s: make(map[scope.Voltage]int),
		}
		for _, d := range s.Samples[ch] {
			hist.s[d]++
		}
		hist.sort()
		if *showHist {
			out := make([]string, len(hist.k))
			for k := range hist.k {
				out[k] = fmt.Sprintf("%f: %d", hist.k[k], hist.s[hist.k[k]])
			}
			fmt.Println(out)
		} else {
			fmt.Println(hist.k[0])
		}
		i -= len(s.Samples[ch])
		if *period != 0 && i <= 0 {
			break
		}
	}
}
