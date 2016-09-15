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

package usbif

import usb "github.com/kylelemons/gousb/usb"

// Device is an interface that mimics usb.Device, but can be replaced for testing
type Device interface {
	Control(rType, request uint8, val, idx uint16, data []byte) (int, error)
	OpenEndpoint(conf, iface, setup, epoint uint8) (usb.Endpoint, error)
	Close() error
	Bus() uint8
	Address() uint8
}

// Desc is a convenience mapping to usb.Descriptor.
type Desc usb.Descriptor

// usbDev is a wrapper around *usb.Device implementing Device interface.
type usbDev struct {
	*usb.Device
}

// Address returns USB device address.
func (d usbDev) Address() uint8 { return d.Device.Address }

// Bus returns USB device bus number.
func (d usbDev) Bus() uint8 { return d.Device.Bus }

// FromRealDevice converts a usb.Device to Device.
func FromRealDevice(d *usb.Device) Device { return usbDev{d} }
