package colorjson

// Settings contains the settings used to colorize and
// format the JSON document.
type Settings struct {
	EndWithNewline bool
	Color          ColorSettings
}

// Default contains the default settings used to colorize
// and format the JSON document.
var Default = Settings{
	EndWithNewline: true,
	Color:          DefaultColors,
}
