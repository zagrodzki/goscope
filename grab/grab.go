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
	"fmt"
	"log"
	"strings"

	"bitbucket.org/zagrodzki/goscope/dummy"
	"bitbucket.org/zagrodzki/goscope/scope"
	"bitbucket.org/zagrodzki/goscope/usb"
)

func must(e error) {
	if e != nil {
		log.Fatalf(e.Error())
	}
}

type system struct {
	enumerate func() map[string]string
	open      func(string) scope.Device
}

var systems = map[string]system{
	"dummy": {
		enumerate: dummy.Enumerate,
		open:      dummy.Open,
	},
	"usb": {
		enumerate: usb.Enumerate,
		open:      usb.Open,
	},
}

func main() {
	var all []string
	for sys := range systems {
		for id := range systems[sys].enumerate() {
			all = append(all, fmt.Sprintf("%s:%s", sys, id))
		}
	}
	if len(all) == 0 {
		log.Fatalf("Did not find any supported devices")
	}
	id := all[0]
	if len(all) > 1 {
		log.Printf("Using the first device (%s)", id)
	}
	parts := strings.SplitN(id, ":", 2)
	osc := systems[parts[0]].open(parts[1])
	fmt.Println(osc)
	for _, ch := range osc.Channels() {
		must(ch.SetVoltRange(5))
	}
	data, _, err := osc.ReadData()
	if err != nil {
		log.Fatalf("ReadData: %+v", err)
	}
	fmt.Println("Data:", data)
}
