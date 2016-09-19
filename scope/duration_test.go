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

func TestDuration(t *testing.T) {
	for _, tc := range []struct {
		t    Duration
		want string
	}{
		{0, "0s"},
		{1, "1fs"},
		{999, "999fs"},
		{1000, "1ps"},
		{1001, "1.001ps"},
		{1100, "1.1ps"},
		{999999, "999.999ps"},
		{1000000, "1ns"},
		{1100000, "1.1ns"},
		{1999999, "1.999999ns"},
		{999999999, "999.999999ns"},
		{1000000000, "1µs"},
		{1100000000, "1.1µs"},
		{999999999999, "999.999999µs"},
		{1000000000000, "1ms"},
		{1000000000001, "1ms"},
		{1999999999999, "1.999999ms"},
	} {
		if got := tc.t.String(); got != tc.want {
			t.Errorf("Duration(%d).String(): got %q, want %q", tc.t, tc.t.String(), tc.want)
		}
	}
}
