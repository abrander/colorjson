package colorjson

import (
	"bytes"
)

// https://seriot.ch/projects/parsing_json.html

// ColorizeData colorizes the given data and returns the result.
func ColorizeData(data []byte, s Settings) ([]byte, error) {
	var buf bytes.Buffer

	e := NewEncoder(&buf, s)
	err := e.Colorize(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Marshal returns the colored JSON encoding of v. If s is nil, the
// default settings are used.
func Marshal(v any, s Settings) ([]byte, error) {
	var buf bytes.Buffer

	err := NewEncoder(&buf, s).Encode(v)

	return buf.Bytes(), err
}
