package dummy

import (
	"testing"

	"bitbucket.org/zagrodzki/goscope/scope"
)

func TestZero(t *testing.T) {
	ch := zeroChan{}
	data := ch.data()
	for _, tc := range []struct {
		idx  int
		want scope.Sample
	}{
		// zero is always 0
		{0, 0},
		{100, 0},
		{999, 0},
	} {
		if got := data[tc.idx]; got != tc.want {
			t.Errorf("zero.data()[%d]: got %v, want %v", tc.idx, got, tc.want)
		}
	}
}
