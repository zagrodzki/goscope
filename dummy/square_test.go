package dummy

import (
	"testing"

	"bitbucket.org/zagrodzki/goscope/scope"
)

func TestSquare(t *testing.T) {
	ch := squareChan{}
	data := ch.data()
	for _, tc := range []struct {
		idx  int
		want scope.Sample
	}{
		// square starts with 1 and flips every 20 cycles.
		{0, 1},
		{10, 1},
		{19, 1},
		{20, -1},
		{21, -1},
		{39, -1},
		{40, 1},
	} {
		if got := data[tc.idx]; got != tc.want {
			t.Errorf("square.data()[%d]: got %v, want %v", tc.idx, got, tc.want)
		}
	}
}
