//  Copyright 2017 The goscope Authors
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

package gui

import "image"

// The lower the alpha channel value, the less visible the pixel should be in
// the resulting blend. We arbitrarily define Alpha of 10 to mean
// "almost invisible" and don't copy these pixels anymore.
const alphaInvisibilityThreshold = 10

// DrawOver copies opaque pixels of src onto dst into the same positions,
// effectively drawing over the dst. DrawOver panics if dst and src are
// not of the same dimensions.
// DrawOver handles the alpha channel only in the most simplified way.
// For complex drawing operations, use image/draw package instead.
func DrawOver(dst *image.RGBA, src *image.RGBA) {
	if dst.Rect != src.Rect {
		panic("DrawOver: dst and src have different bounds.")
	}
	// every pixel is represented by four consecutive bytes in Pix:
	// red, green, blue, alpha.
	for i := 0; i < len(dst.Pix); i += 4 {
		if src.Pix[i+3] < alphaInvisibilityThreshold {
			continue
		}
		copy(dst.Pix[i:i+4], src.Pix[i:i+4])
	}
}
