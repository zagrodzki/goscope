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
