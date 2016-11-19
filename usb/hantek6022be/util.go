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

package hantek6022be

import (
	"fmt"
	"math"
	"strconv"
)

// format a number, with K/M/G suffix and limiting the precision to 3 digits.
func fmtVal(v float64) string {
	av := math.Abs(v)
	sfx := ""
	switch {
	case av >= 999999500:
		v /= 1e9
		sfx = "G"
	case av >= 999999.5:
		v /= 1e6
		sfx = "M"
	case av >= 999.9995:
		v /= 1e3
		sfx = "K"
	}
	ret := strconv.FormatFloat(v, 'f', 3, 64)
	for ret[len(ret)-1] == '0' {
		ret = ret[:len(ret)-1]
	}
	if ret[len(ret)-1] == '.' {
		ret = ret[:len(ret)-1]
	}
	return fmt.Sprintf("%s%s", ret, sfx)
}
