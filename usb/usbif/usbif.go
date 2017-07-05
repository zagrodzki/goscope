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
	"github.com/google/gousb"
	"github.com/pkg/errors"
)

// Device is an interface that mimics gousb.Device, but can be replaced for testing
type Device interface {
	Control(rType, request uint8, val, idx uint16, data []byte) (int, error)
	OpenEndpoint(conf, iface, setup, epoint int) (*gousb.InEndpoint, error)
	Close() error
	Bus() int
	Address() int
	Configs() map[int]gousb.ConfigDesc
}

// usbDev is a wrapper around *gousb.Device implementing Device interface.
type usbDev struct {
	*gousb.Device
	conf  *gousb.Config
	iface *gousb.Interface
}

// Address returns USB device address.
func (d usbDev) Address() int { return d.Device.Desc.Address }

// Bus returns USB device bus number.
func (d usbDev) Bus() int { return d.Device.Desc.Bus }

// Configs returns a list of available USB device configs.
func (d usbDev) Configs() map[int]gousb.ConfigDesc {
	return d.Device.Desc.Configs
}

// OpenEndpoint is a wrapper that sets the device config, claims the interface
// and returns an InEndpoint ready for read.
func (d usbDev) OpenEndpoint(conf, iface, setup, epoint int) (*gousb.InEndpoint, error) {
	if d.conf != nil {
		return nil, errors.Errorf("OpenEndpoint(...): device %s already has a config claimed, two endpoints at the same time are not supported.", d)
	}
	c, err := d.Config(conf)
	if err != nil {
		return nil, errors.Wrapf(err, "(%s).Config(%d) failed: %v", d, conf, err)
	}
	i, err := c.Interface(iface, setup)
	if err != nil {
		c.Close()
		return nil, errors.Wrapf(err, "(%s).Config(%d).Interface(%d, %d) failed: %v", d, conf, iface, setup, err)
	}
	ep, err := i.InEndpoint(epoint)
	if err != nil {
		i.Close()
		c.Close()
		return nil, errors.Wrapf(err, "(%s).Config(%d).Interface(%d, %d).InEndpoint(%d) failed: %v", d, conf, iface, setup, epoint, err)
	}
	d.conf = c
	d.iface = i
	return ep, nil
}

// Close releases the device and any config/interface it may have claimed.
func (d usbDev) Close() error {
	var err error
	if d.iface != nil {
		d.iface.Close()
		d.iface = nil
	}
	if d.conf != nil {
		err = d.conf.Close()
		d.conf = nil
	}
	err = d.Close()
	return err
}

// FromRealDevice converts a gousb.Device to Device.
func FromRealDevice(d *gousb.Device) Device {
	return usbDev{
		Device: d,
	}
}
