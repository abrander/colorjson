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
	s Settings

	useColors    bool
	currentColor Color
	state        state
	newline1     bool
	newline2     bool
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer, s Settings) *Encoder {
	return &Encoder{
		w:         w,
		s:         s,
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

	readThing := func(r rune) {
		if r < 0 {
			return
		}

		switch r {
		case '{':
			color = Reset
			e.newline2 = true
			e.state.push(object)

		case '[':
			color = Reset
			e.newline2 = true
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
				color = Reset
				e.newline1 = true
				e.state.pop()
			}

			if r == ',' {
				color = Reset
				e.newline2 = true
			}

			if r == '"' {
				color = e.s.Color.Ident
				e.state.push(identifier)

				printRune(r)
				readString(postIdentifier)

				break
			}

			printRune(r)

		case array:
			r := skipSpace()
			switch r {
			case ']':
				color = Reset
				e.newline1 = true
				e.state.pop()

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
