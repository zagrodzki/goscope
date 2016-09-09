package oscilloscope

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

type ChanID string

type Channel interface {
	ID() ChanID
	SetVoltRange(float64) error
}

type VoltRange float64

func (v VoltRange) String() {
	return fmt.Sprintf("+-%fV", v)
}

type SampleRate int

func fmtVal(v float64) string {
	av := math.Abs(v)
	sfx := ""
	switch {
	case av >= 1e9:
		v /= 1e9
		sfx = "G"
	case av >= 1e6:
		v /= 1e6
		sfx = "M"
	case av >= 1e3:
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

func (s SampleRate) String() {
	return fmt.Sprintf("%s samples/s", fntVal(float64(s)))
}

type Device interface {
	String() string
	Channels() map[ChanID]Channel
	StartCapture() error
	StopCapture() error
	ReadData() (map[ChanID][]byte, time.Duration, error)
}
