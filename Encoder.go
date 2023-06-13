package colorjson

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// Encoder is a JSON encoder that colorizes the output.
type Encoder struct {
	w io.Writer
	s *Settings
	c *ColorSettings

	useColors    bool
	currentColor Color
	state        state
	newline1     bool
	newline2     bool
}

// NewEncoder returns a new encoder that writes to w. If settings
// is nil, the default settings are used.
func NewEncoder(w io.Writer, settings *Settings) *Encoder {
	s := Default
	c := DefaultColors

	if settings != nil {
		s = settings

		if settings.Color != nil {
			c = settings.Color
		}
	}

	return &Encoder{
		w:         w,
		s:         s,
		c:         c,
		useColors: s.ColorMode.UseColors(w),
	}
}

// Encode writes the JSON encoding of v to the stream.
func (e *Encoder) Encode(v any) error {
	return json.NewEncoder(e).Encode(v)
}

// Write writes p to the stream. The function expects p to contain
// valid JSON data.
func (e *Encoder) Write(p []byte) (n int, err error) {
	return len(p), e.ColorizeData(p)
}

// ColorizeData colorizes the JSON data and writes it to the stream.
func (e *Encoder) ColorizeData(data []byte) error {
	return e.Colorize(bytes.NewReader(data))
}

// Colorize reads JSON from r and writes colorized JSON to the stream.
func (e *Encoder) Colorize(r io.Reader) error {
	var err error

	const ERROR rune = -1

	reader := bufio.NewReader(r)

	readOne := func() rune {
		var r rune

		r, _, err = reader.ReadRune()
		if err != nil {
			return ERROR
		}

		return r
	}

	skipSpace := func() rune {
		r := ' '

		for unicode.IsSpace(r) {
			r = readOne()
		}

		return r
	}

	color := Reset

	printConditionalNewline := func(needed *bool) {
		if !*needed {
			return
		}

		_, _ = e.w.Write([]byte("\n"))
		if e.s.Indent != "" {
			for i := 0; i < e.state.depth(); i++ {
				_, _ = e.w.Write([]byte(e.s.Indent))
			}
		}

		*needed = false
	}

	printRune := func(r rune) {
		if err != nil {
			return
		}

		if e.useColors && color != e.currentColor {
			e.currentColor = color

			_, _ = e.w.Write([]byte(color.String()))
		}

		printConditionalNewline(&e.newline1)
		_, err = e.w.Write([]byte(string(r)))
		printConditionalNewline(&e.newline2)
	}

	stateColor := func(r rune) (int, Color, error) {
		switch r {
		case '{':
			return objectStart, Reset, nil

		case '[':
			return arrayStart, Reset, nil

		case '"':
			return stringValue, e.c.String, nil

		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-', '+':
			return numberValue, e.c.Number, nil

		case 't':
			return boolValue, e.c.True, nil

		case 'f':
			return boolValue, e.c.False, nil

		case 'n':
			return nullValue, e.c.Null, nil

		default:
			return start, Reset, fmt.Errorf("Unexpected character: %c", r)
		}
	}

	readThing := func(r rune) {
		if r < 0 {
			return
		}

		s, c, e2 := stateColor(r)
		if e2 != nil {
			err = e2
		}

		color = c
		e.state.push(s)

		printRune(r)
	}

	readString := func(nextState int) {
		var r rune

		for r != '"' {
			r = readOne()
			if r < 0 {
				break
			}

			if r == '\\' {
				printRune(r)

				r = readOne()
				if r < 0 {
					break
				}
				printRune(r)

				continue
			}

			if r == '"' {
				if nextState < 0 {
					e.state.pop()
				} else {
					e.state.replace(nextState)
				}
				printRune(r)

				color = Reset
			} else {
				printRune(r)
			}

		}
	}

READLOOP:
	for err == nil {
		state := e.state.state()
		switch state {
		case start:
			readThing(skipSpace())

		case objectStart, object:
			r := skipSpace()
			if r < 0 {
				break READLOOP
			}

			if r == '}' {
				color = Reset
				if state != objectStart {
					e.newline1 = true
				}
				e.state.pop()
			}

			if r == ',' {
				color = Reset
				e.newline2 = true
			}

			if r == '"' {
				color = e.c.Ident

				if state == objectStart {
					e.newline1 = true
					e.state.replace(object)
				}

				printRune(r)
				e.state.push(identifier)

				readString(postIdentifier)

				break
			}

			printRune(r)

		case arrayStart, array:
			r := skipSpace()
			switch r {
			case ']':
				color = Reset
				if state != arrayStart {
					e.newline1 = true
				}
				e.state.pop()

				printRune(r)

			case ',':
				color = Reset
				e.newline2 = true
				printRune(r)

			case ERROR:
				break READLOOP

			default:
				color = e.c.Ident
				if state == arrayStart {
					e.newline1 = true
					e.state.replace(array)
				}

				s, c, _ := stateColor(r)
				color = c
				printRune(r)
				e.state.push(s)
			}

		case postIdentifier:
			r := skipSpace()
			if r < 0 {
				break READLOOP
			}

			if r == ':' {
				e.state.replace(preValue)
				color = Reset
			}

			printRune(r)

		case preValue:
			_, err = e.w.Write([]byte(e.s.Separator))

			e.state.pop()
			readThing(skipSpace())

		case stringValue:
			readString(-1)

		case numberValue:
			r := readOne()
			if r < 0 {
				break READLOOP
			}

			switch r {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-', '+', '.', 'e', 'E':
				printRune(r)

			case ',':
				color = Reset
				e.newline2 = true
				e.state.pop()

				printRune(r)

			case ']', '}':
				color = Reset
				e.newline1 = true
				e.state.pop() // exit numberValue state
				e.state.pop() // exit array or object state

				printRune(r)

			default:
				color = Red | Bold

				if !unicode.IsSpace(r) {
					printRune(r)
				}
			}

		case nullValue, boolValue:
			r := readOne()
			if r < 0 {
				break READLOOP
			}

			if strings.ContainsRune("null"+"true"+"false", r) {
				printRune(r)

				break
			}

			if r == ',' {
				color = Reset
				e.newline2 = true
				e.state.pop()
			}

			if r == '}' || r == ']' {
				color = Reset
				e.newline1 = true
				e.state.pop() // end the value
				e.state.pop() // end the surrounding object/array
			}

			if !unicode.IsSpace(r) {
				printRune(r)
			}
		}
	}

	if err == io.EOF {
		err = nil

		if e.s.EndWithNewline {
			_, err = e.w.Write([]byte{'\n'})
		}
	}

	return err
}
