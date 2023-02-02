package colorjson

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"unicode"
)

// Encoder is a JSON encoder that colorizes the output.
type Encoder struct {
	w io.Writer
	s Settings

	currentColor Color
	state        state
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer, s Settings) *Encoder {
	return &Encoder{
		w: w,
		s: s,
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

	printRune := func(r rune) {
		if err != nil {
			return
		}

		if color != e.currentColor {
			e.currentColor = color

			_, _ = e.w.Write([]byte(color.String()))
		}

		_, err = e.w.Write([]byte(string(r)))
	}

	readThing := func(r rune) {
		if r < 0 {
			return
		}

		switch r {
		case '{':
			color = Reset
			e.state.push(object)

		case '[':
			color = Reset
			e.state.push(array)

		case '"':
			color = e.s.Color.String
			e.state.push(stringValue)

		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-', '+':
			color = e.s.Color.Number
			e.state.push(numberValue)

		case 't':
			color = e.s.Color.True
			e.state.push(boolValue)

		case 'f':
			color = e.s.Color.False
			e.state.push(boolValue)

		case 'n':
			color = e.s.Color.Null
			e.state.push(nullValue)

		default:
			err = fmt.Errorf("Unexpected character: %c", r)
		}

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
		switch e.state.state() {
		case start:
			readThing(skipSpace())

		case object:
			r := skipSpace()
			if r < 0 {
				break READLOOP
			}

			if r == '}' {
				e.state.pop()
				color = Reset
			}

			if r == ',' {
				color = Reset
			}

			if r == '"' {
				e.state.push(identifier)
				color = e.s.Color.Ident

				printRune(r)
				readString(postIdentifier)

				break
			}

			printRune(r)

		case array:
			r := skipSpace()
			switch r {
			case ']':
				e.state.pop()

				color = Reset
				printRune(r)

			case ',':
				color = Reset
				printRune(r)

			case ERROR:
				break READLOOP

			default:
				readThing(r)
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
			e.state.pop()
			readThing(skipSpace())

		case stringValue:
			readString(-1)

		case numberValue:
			r := readOne()
			if r < 0 {
				break READLOOP
			}

			if !unicode.IsDigit(r) && r != '.' && r != 'e' && r != 'E' && r != '+' && r != '-' {
				color = Reset

				e.state.pop()
			}

			if r == ']' || r == '}' {
				color = Reset

				e.state.pop()
			}

			if !unicode.IsSpace(r) {
				printRune(r)
			}

		case nullValue, boolValue:
			r := readOne()
			if r < 0 {
				break READLOOP
			}

			if r == ',' || r == '}' || r == ']' || unicode.IsSpace(r) {
				color = Reset

				e.state.pop()
			}

			if !unicode.IsSpace(r) {
				printRune(r)
			}
		}
	}

	if err == io.EOF && e.s.EndWithNewline {
		_, err = e.w.Write([]byte{'\n'})
	}

	return err
}
