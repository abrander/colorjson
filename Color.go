package colorjson

import (
	"fmt"
)

// Color is a color used to highlight something in the JSON document.
type Color int

const (
	Reset   Color = 0
	Grey    Color = 30 | Bold
	Red     Color = 31
	Green   Color = 32
	Yellow  Color = 33
	Blue    Color = 34
	Magenta Color = 35
	Cyan    Color = 36
	White   Color = 37
)

const (
	Bold Color = 0x100
)

func (c Color) String() string {
	res := "\033["

	if c&Bold != 0 {
		res += "1;"
	}

	c &= ^Bold

	return res + fmt.Sprintf("%dm", c)
}
