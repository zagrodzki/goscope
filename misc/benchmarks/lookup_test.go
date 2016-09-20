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

package benchmark

import "testing"

type lookup func(string) int

func lookupMap(x string) int {
	return map[string]int{
		"foo": 1,
		"bar": 2,
		"baz": 3,
	}[x]
}

func lookupSwitch(x string) int {
	switch x {
	case "foo":
		return 1
	case "bar":
		return 2
	case "baz":
		return 3
	}
	return 0
}

func BenchmarkLookup(b *testing.B) {
	var out int
	for _, bc := range []struct {
		name string
		f    lookup
	}{
		{"map", lookupMap},
		{"switch", lookupSwitch},
	} {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				out = bc.f("foo")
			}
		})
	}
	b.Logf("%d", out)
}
