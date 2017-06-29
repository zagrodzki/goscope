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

import (
	"errors"
	"fmt"

	usb "github.com/kylelemons/gousb/usb"
)

// Device is an interface that mimics usb.Device, but can be replaced for testing
type Device interface {
	Control(rType, request uint8, val, idx uint16, data []byte) (int, error)
	InEndpoint(ifNum, alt, epNum int) (*usb.InEndpoint, error)
	Close() error
	Bus() int
	Address() int
	Configs() map[int]usb.ConfigInfo
	Config() (usb.ConfigInfo, error)
	SetConfig(int) error
}

// Config is the device with selected active config.

// Desc is a convenience mapping to usb.Descriptor.
type Desc usb.Descriptor

// usbDev is a wrapper around *usb.Device implementing Device interface.
type usbDev struct {
	d *usb.Device
	c *usb.Config
	i *usb.Interface
}

// Address returns USB device address.
func (d *usbDev) Address() int { return d.d.Address }

// Bus returns USB device bus number.
func (d *usbDev) Bus() int { return d.d.Bus }

// Configs returns a list of available USB device configs.
func (d *usbDev) Configs() map[int]usb.ConfigInfo { return d.d.Descriptor.Configs }

// Config returns the current config.
func (d *usbDev) Config() (usb.ConfigInfo, error) {
	if d.c == nil {
		return usb.ConfigInfo{}, errors.New("no active config, SetConfig not called yet")
	}
	return d.c.Info, nil
}

// Control sends a control transfer.
func (d *usbDev) Control(rType, request uint8, val, idx uint16, data []byte) (int, error) {
	return d.c.Control(rType, request, val, idx, data)
}

// InEndpoint returns a new InEndpoint from the given interface in current configuration.
func (d *usbDev) InEndpoint(ifNum, alt, epNum int) (*usb.InEndpoint, error) {
	intf, err := d.c.Interface(ifNum, alt)
	if err != nil {
		return nil, err
	}
	d.i = intf
	return intf.InEndpoint(epNum)
}

// Close releases the interface and configuration, then closes the device.
func (d *usbDev) Close() error {
	if d.i != nil {
		d.i.Close()
		d.i = nil
	}
	if d.c != nil {
		if err := d.c.Close(); err != nil {
			return err
		}
		d.c = nil
	}
	return d.d.Close()
}

// SetConfig activates a device configuration.
func (d *usbDev) SetConfig(cfgNum int) error {
	if d.c != nil {
		return fmt.Errorf("SetConfig(%d): USB device %s already has an active config %d", d.d, d.c.Info.Config)
	}
	cfg, err := d.d.Config(cfgNum)
	if err != nil {
		return err
	}
	d.c = cfg
	return nil
}

// FromRealDevice converts a usb.Device to Device.
func FromRealDevice(d *usb.Device) Device { return &usbDev{d: d} }
