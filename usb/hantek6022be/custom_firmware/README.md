This is a custom firmware from Hantek6022BE, written by jhoenicke (https://github.com/jhoenicke).
It's largely compatible with the original firmware delivered by the manufacturer, but with important differences:

* supports transfers in isochronous mode on endpoint 2 of EZ-USB controller. This provides large USB buffer and reserved bandwidth on the USB bus.
* supports a single-channel mode, wich allows higher capture rate for the channel.[1]
* doesn't support EEPROM access for reading the calibration data.
* front panel LED is much easier to read - it's green when device is active and ready to send data. When stopping sampling, the LED turns red for a short moment and then turns dark.

Our copy of the firmware code is modified through enabling auto-adjustment of the isochronous packet size, which jhoenicke's firmware didn't have as of 2017.02.05.

[1] the highest isochronous throughput on the high-speed USB bus is a 24.5MiB/s (every 125ms frame contains 3 packets of 1024 bytes).
    Using bulk transfers practical throughput is somewhere above 40MiB/s, but in practice you will see gaps in captured data, which is why isochronous transfers are preferred for streaming.
    Each sample is one byte, which limits sampling rate for dual-channel mode to 12Msps.
    Disabling one channel allows effective doubling of sampling rate to 24Msps on the single active channel.

=== Installation

For your convenience a pre-compiled version of the firmware is included in `firmware.hex`.

Copy `firmware.hex` to `/usr/local/share/hantek` as `hantek6022be-custom.hex`. Also copy the second stage loader, `hantek6022be-loader.hex`, to the same place.
You can get the stock loader (and stock firmware) from here: https://github.com/olerem/openhantek/tree/6022be/fw

Install fxload.

In udev rules, add:

    SUBSYSTEM=="usb", ACTION=="add", ENV{DEVTYPE}=="usb_device", ENV{PRODUCT}=="4b4/6022/*", RUN+="/sbin/fxload -t fx2 -I /usr/local/share/hantek/hantek6022be-custom.hex -s /usr/local/share/hantek/hantek6022be-loader.hex -D $env{DEVNAME}"
    ATTRS{idVendor}=="04b4", ATTRS{idProduct}=="6022", MODE="0660", GROUP="plugdev"

reload udev rules with:

    udevadm control --reload

replug the device and you should be good to go.

=== Bulding

You'll need sdcc installed. Run `make` to compile the firmware. Resulting file is in `build/firmware.ihx`. Convert it to the hex format using:

    packihx build/firmware.ihx > firmware.hex
