package scope

// A control element of the Device - a button, knob, switch on the scope front panel.
type Param interface {
	// Name returns the name of the param, intended for the UI.
	Name() string
	// Value returns the current value displayed as string, for the UI.
	Value() string
	// Set sets a new value of the param. If the value is invalid, a non-nil error is returned.
	Set(string) error
}

// Anything that would be represented by a button or a discrete knob/slider. This control
// represents discrete settings that don’t have an ordering - e.g. name, type etc.
// or a range with not many discrete values (e.g. time base length). For ranges with many
// values (unsuitable for a drop-down list), a RangeParam is better.
type SelectParam interface {
	Param
	// Values returns the list of available values to choose from.
	Values() []string
}

// A continuous control knob/slider. The speed of change should accelerate when the user
// keeps increasing the value at a quick steady pace.
type RangeParam interface {
	Param
	// Inc increases the param setting.
	Inc() string
	// Dec decreases the param setting.
	Dec() string
}
