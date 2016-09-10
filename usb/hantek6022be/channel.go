package hantek6022be

import (
	"fmt"

	"bitbucket.org/zagrodzki/goscope/scope"
	"github.com/pkg/errors"
)

type ch struct {
	id        scope.ChanID
	osc       *Scope
	voltRange scope.VoltRange
}

func (c ch) ID() scope.ChanID { return c.id }
func (ch) GetVoltRanges() []scope.VoltRange {
	return voltRanges
}
func (c ch) GetVoltRange() scope.VoltRange {
	return c.voltRange
}
func (c ch) SetVoltRange(v scope.VoltRange) error {
	req := map[scope.ChanID]uint8{
		"CH1": ch1VoltRangeReq,
		"CH2": ch2VoltRangeReq,
	}[c.id]
	val, ok := voltRangeToID[v]
	if !ok {
		return errors.New(fmt.Sprintf("Channel %s: SetVoltRange(%s): range must be one of %v", c, v, voltRanges))
	}
	if _, err := c.osc.dev.Control(0x40, req, 0, 0, val.data()); err != nil {
		return errors.Wrapf(err, "Control(voltage range %s(%x))", v, val)
	}
	c.voltRange = v
	return nil
}
