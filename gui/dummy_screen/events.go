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

package main

import (
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
)

// wait for image events, like mouse click, key press etc.
func processEvents(eq screen.EventDeque, stop chan<- struct{}) {
	done := false
	for {
		e := eq.NextEvent()
		switch v := e.(type) {
		case lifecycle.Event:
			if v.To == lifecycle.StageDead {
				done = true
			}
		case key.Event:
			if v.Code == key.CodeEscape || (v.Code == key.CodeC && v.Modifiers&key.ModControl > 0) {
				done = true
			}
		}
		if done {
			stop <- struct{}{}
			return
		}
	}
}
