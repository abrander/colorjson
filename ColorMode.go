package colorjson

import (
	"io"
)

// ColorMode is used to determine when to colorize the output.
type ColorMode int

const (
	// Never disables colorization.
	Never ColorMode = -1

	// Auto enables colorization if the output is a terminal.
	Auto ColorMode = 0

	// Always enables colorization.
	Always ColorMode = 1
)

// UseColors returns true if the output should be colorized.
func (c ColorMode) UseColors(w io.Writer) bool {
	switch c {
	case Always:
		return true

	case Never:
		return false

	default:
		return isatty(w)
	}
}
