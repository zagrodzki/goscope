package scope

// TraceParams represents various trace parameters related to the display.
type TraceParams struct {
	// the position of Y=0 (0 <= Zero <= 1) given as
	// the fraction of the window height counting from the bottom
	Zero float64
	// volts per div
	PerDiv float64
}
