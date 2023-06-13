package colorjson

// ColorSettings contains the colors used to highlight the different
// parts of the JSON document.
type ColorSettings struct {
	Ident  Color
	String Color
	Number Color
	Null   Color
	False  Color
	True   Color
}

// JqColors is the color scheme used by jq.
var JqColors = &ColorSettings{
	Ident:  Blue | Bold,
	String: Green,
	Number: Reset,
	Null:   Grey,
	False:  Reset,
	True:   Reset,
}

// DefaultColors is the default color scheme.
var DefaultColors = &ColorSettings{
	Ident:  Cyan | Bold,
	String: Green,
	Number: Yellow,
	Null:   Grey,
	False:  Red,
	True:   Green,
}
