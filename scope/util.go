package scope

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
