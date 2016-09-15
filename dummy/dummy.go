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
