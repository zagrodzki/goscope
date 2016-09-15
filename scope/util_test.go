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

import "testing"

func TestFmtVal(t *testing.T) {
	for _, tc := range []struct {
		v    float64
		want string
	}{
		{0, "0"},
		{123, "123"},
		{123.3, "123.3"},
		{123.43, "123.43"},
		{123.543, "123.543"},
		{123.6543, "123.654"},
		{123.6666, "123.667"},
		{999.99, "999.99"},
		{999.999, "999.999"},
		{999.9999, "1K"},
		{1023.3, "1.023K"},
		{11234.5, "11.235K"},
		{999999.999, "1M"},
		{1000000, "1M"},
		{1000001, "1M"},
		{11000000, "11M"},
		{666666666, "666.667M"},
		{9000111111, "9G"},
	} {
		if got := fmtVal(tc.v); got != tc.want {
			t.Errorf("fmtVal(%f): got %q, want %q", tc.v, got, tc.want)
		}
	}
}
