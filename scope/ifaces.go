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

package scope

import "image"

// TimeWindow defines a period of time.
type TimeWindow interface {
	TimeBase() Duration
}

// DataRecorder stores data from the device.
type DataRecorder interface {
	TimeWindow
	// Reset clears the DataRecorder and sets it's parameters.
	Reset(Duration, <-chan []ChannelData)
	// Error registers an error occurence.
	// After an error, no more data will be written to the recorder.
	Error(error)
}

// DisplayLayer represents partial data on the osciloscope display.
type DisplayLayer interface {
	TimeWindow
	// SetTimeBase sets the length of the recorded sweeps.
	SetTimeBase(Duration)

	// SetChannel configures the display parameters for an input channel.
	// ScreenPos is in the range of 0..1 where 0 represents the bottom of the screen,
	// and 1 represents the very top of the screen.
	SetChannel(ch ChanID, screenPos float32, perDiv Voltage)

	// Render returns some image data. An image can and typically should be transparent.
	// A nil image means the image did not change from last Draw() call.
	Render() *image.RGBA
}
