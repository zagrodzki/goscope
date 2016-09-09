package hantek6022be

import (
	"fmt"

	"github.com/pkg/errors"
	"zagrodzki.net/gohantek/oscilloscope"
)

const (
	ch1VoltRangeReq uint8 = 0xe0
	ch2VoltRangeReq uint8 = 0xe1
)

type ch struct {
	id  oscilloscope.ChanID
    voltRange float64
	osc *osc
}

type rangeID uint8

func (v rangeID) Data() []byte {
	return []byte{byte(v)}
}

var voltRanges = []float64{0.5, 1, 2.5, 5}
var voltRangeToID = map[float64]rangeID{
	5:   0x01,
	2.5: 0x02,
	1:   0x05,
	0.5: 0x0a,
}

func (c ch) ID() oscilloscope.ChanID { return c.id }
func (c ch) GetVoltRange() float64 {
    return c.voltRange
}
func (c ch) SetVoltRange(v float64) error {
	req := map[oscilloscope.ChanID]uint8{
		"CH1": ch1VoltRangeReq,
		"CH2": ch2VoltRangeReq,
	}[c.id]
	val, ok := voltRangeMap[v]
	if !ok {
		return errors.New(fmt.Sprintf("Channel %s: SetVoltRange(%f): range must be one of %v", c, v, voltRanges))
	}
	if _, err := c.osc.dev.Control(0x40, req, 0, 0, val.Data()); err != nil {
            return errors.Wrapf(err, "Control(voltage range %f(%x))", v, val)
    }
    c.voltRange = v
    return nil
}
