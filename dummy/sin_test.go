package dummy

import (
	"testing"

	"bitbucket.org/zagrodzki/goscope/scope"
)

const epsilon = 0.01

func almostEqual(a, b scope.Sample) bool {
	return (a-b) < epsilon && (b-a) < epsilon
}

func TestSin(t *testing.T) {
	ch := sinChan{}
	data := ch.data()
	for _, tc := range []struct {
		idx  int
		want scope.Sample
	}{
		// sin is a sine wave with a period of 10pi. Values are approximate to .01.
		{0, 0},
		{8, 1},
		{13, 0.516},
		{55, -1},
	} {
		if got := data[tc.idx]; !almostEqual(got, tc.want) {
			t.Errorf("sin.data()[%d]: got %v, want %v", tc.idx, got, tc.want)
		}
	}
}
