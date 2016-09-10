package dummy

import (
	"testing"

	"bitbucket.org/zagrodzki/goscope/scope"
)

func TestTriangle(t *testing.T) {
	ch := triangleChan{}
	data := ch.data()
	for _, tc := range []struct {
		idx  int
		want scope.Sample
	}{
		// triangle starts with -1, goes to 1 over 20 cycles and goes back to -1 over another 20.
		{0, -1},
		{1, -0.9},
		{10, 0},
		{20, 1},
		{21, 0.9},
		{30, 0},
		{40, -1},
	} {
		if got := data[tc.idx]; got != tc.want {
			t.Errorf("triangle.data()[%d]: got %v, want %v", tc.idx, got, tc.want)
		}
	}
}
