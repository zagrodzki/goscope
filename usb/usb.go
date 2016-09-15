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

// Package usb contains discovery routines to enumerate supported devices connected via USB.
package usb

import (
	"fmt"
	"log"

	"bitbucket.org/zagrodzki/goscope/scope"
	"bitbucket.org/zagrodzki/goscope/usb/hantek6022be"
	"bitbucket.org/zagrodzki/goscope/usb/usbif"
	"github.com/kylelemons/gousb/usb"
)

type driver struct {
	name  string
	check func(*usbif.Desc) bool
	open  func(usbif.Device) scope.Device
}

var drivers = []driver{
	{
		name:  "Hantek 6022BE",
		check: hantek6022be.SupportsUSB,
		open:  func(d usbif.Device) scope.Device { return hantek6022be.New(d) },
	},
}

// connectedDev stores information about identified device
type connectedDev struct {
	// bus and addr copied from USB descriptor
	bus  uint8
	addr uint8
	// driver is an index to the drivers slice
	driver int
}

// String returns an identification of a connected device in a human readable form.
func (d connectedDev) String() string {
	return fmt.Sprintf("%s at USB bus %d addr %d", drivers[d.driver].name, d.bus, d.addr)
}

// found keeps all the connected devices found during enumeration.
var found map[string]connectedDev

// Enumerate finds all connected devices and returns their list. The device
// number can be later used to open a device.
func Enumerate() map[string]string {
	ctx := usb.NewContext()
	found = make(map[string]connectedDev)
	_, err := ctx.ListDevices(func(d *usb.Descriptor) bool {
		for i, s := range drivers {
			if s.check((*usbif.Desc)(d)) {
				newDev := connectedDev{
					bus:    d.Bus,
					addr:   d.Address,
					driver: i,
				}
				fmt.Println("Found:", newDev)
				found[fmt.Sprintf("%d:%d", d.Bus, d.Address)] = newDev
				return false
			}
		}
		return false
	})
	if err != nil {
		log.Printf("ctx.ListDevices(): %v", err)
		return nil
	}
	ret := make(map[string]string)
	for id, val := range found {
		ret[id] = val.String()
	}
	return ret
}

// Open opens a device using an index that was earlier returned from Enumerate()
// After the scope is no longer in use, the caller must call it's Close() method.
func Open(s string) scope.Device {
	dev, ok := found[s]
	if !ok {
		log.Fatalf("Device %s was not found in the enumerated list. Available devices: %v", s, found)
	}
	ctx := usb.NewContext()
	usbDev, err := ctx.ListDevices(func(d *usb.Descriptor) bool {
		return d.Address == dev.addr && d.Bus == dev.bus
	})
	if err != nil {
		log.Fatalf("ctx.ListDevices(): %v", err)
	}
	if len(usbDev) != 1 {
		log.Fatalf("Expected exactly 1 device to be open after ctx.ListDevices, got %d", len(usbDev))
	}
	desc := (*usbif.Desc)(usbDev[0].Descriptor)
	if !drivers[dev.driver].check(desc) {
		log.Fatalf("%s check() on the usb device %d:%d (vendor/product %04x:%04x) unexpectedly returned false", drivers[dev.driver].name, desc.Bus, desc.Address, desc.Vendor, desc.Product)
	}
	return drivers[dev.driver].open(usbif.FromRealDevice(usbDev[0]))
}
