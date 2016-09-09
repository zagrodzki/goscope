package main

import (
	"fmt"
	"log"

	"bitbucket.org/zagrodzki/goscope/hantek6022be"
	"bitbucket.org/zagrodzki/goscope/scope"
	"github.com/kylelemons/gousb/usb"
)

type supportedModel struct {
	check func(*usb.Descriptor) bool
	open  func(*usb.Device) scope.Device
}

var supportedModels = []supportedModel{
	supportedModel{hantek6022be.SupportsUSB, hantek6022be.NewScope},
}

func isSupported(d *usb.Descriptor) bool {
	for _, s := range supportedModels {
		if s.check(d) {
			return true
		}
	}
	return false
}

func open(d *usb.Device) scope.Device {
	for _, s := range supportedModels {
		if s.check(d.Descriptor) {
			return s.open(d)
		}
	}
	return nil
}

func must(e error) {
	if e != nil {
		log.Fatalf(e.Error())
	}
}

func main() {
	ctx := usb.NewContext()
	devices, err := ctx.ListDevices(isSupported)
	defer func() {
		for _, d := range devices {
			d.Close()
		}
	}()
	if err != nil {
		log.Fatalf("ctx.ListDevices(): %v", err)
	}
	if len(devices) == 0 {
		log.Fatal("Did not find a valid device")
	}
	for _, d := range devices {
		fmt.Printf("Device found at bus %d addr %d\n", d.Bus, d.Address)
	}
	if len(devices) > 1 {
		fmt.Println("Using the first device listed")
	}
	osc := open(devices[0])
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
