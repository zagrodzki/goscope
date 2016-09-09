package main

import (
	"fmt"
	"log"

	"github.com/kylelemons/gousb/usb"
)

const (
	Hantek6022BEVendor  = 0x4b5
	Hantek6022BEProduct = 0x6022
)

func main() {
	ctx := usb.NewContext()
	devices, err := ctx.ListDevices(func(d *usb.Descriptor) bool {
		return d.Vendor == Hantek6022BEVendor && d.Product == Hantek6022BEProduct
	})
	defer func() {
		for _, d := range devices {
			d.Close()
		}
	}()
	if err != nil {
		log.Fatalf("ctx.ListDevices(): %v", err)
	}
	for _, d := range devices {
		fmt.Printf("Device found at bus %d addr %d\n", d.Bus, d.Address)
	}
}
