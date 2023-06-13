package colorjson

// Settings contains the settings used to colorize and
// format the JSON document.
type Settings struct {
	EndWithNewline bool
	Newlines       bool
	Indent         string
	Separator      string
	ColorMode      ColorMode
	Color          *ColorSettings
}

// Default contains the default settings used to colorize
// and format the JSON document.
var Default = &Settings{
	EndWithNewline: true,
	Newlines:       true,
	Indent:         "  ",
	Separator:      " ",
	Color:          nil,
}

// Jq contains the settings used to colorize and
// format the JSON document to look like the output
// from jq.
var Jq = &Settings{
	EndWithNewline: true,
	Newlines:       true,
	Indent:         "  ",
	Separator:      " ",
	Color:          JqColors,
}
