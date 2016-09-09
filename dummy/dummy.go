package dummy

import (
	"log"

	"bitbucket.org/zagrodzki/goscope/scope"
)

const numSamples = 1000

// Enumerate returns the one and only dummy device
func Enumerate() map[string]string {
	log.Printf("Found: a dummy device")
	return map[string]string{
		"dummy": "a dummy capture device",
	}
}

// Open opens the dummy device
func Open(string) scope.Device {
	return dum{}
}
