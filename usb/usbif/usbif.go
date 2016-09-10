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

// usbDev is a wrapper around *usb.Device implementing Device interface.
type usbDev struct {
	*usb.Device
}

// Address returns USB device address.
func (d usbDev) Address() uint8 { return d.Device.Address }

// Bus returns USB device bus number.
func (d usbDev) Bus() uint8 { return d.Device.Bus }

// FromLibUSBDevice converts a usb.Device to Device
func FromRealDevice(d *usb.Device) usbDev {
	return usbDev{d}
}

// Desc is a convenience mapping to usb.Descriptor
type Desc usb.Descriptor
