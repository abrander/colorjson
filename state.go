package colorjson

import (
	"fmt"
)

type state struct {
	prev    []int
	current int
}

const (
	start = iota
	objectStart
	object
	arrayStart
	array
	identifier
	postIdentifier
	preValue
	stringValue
	numberValue
	boolValue
	nullValue
)

func (s *state) state() int {
	return s.current
}

func (s *state) pop() {
	if len(s.prev) == 0 {
		return
	}

	s.current = s.prev[len(s.prev)-1]
	s.prev = s.prev[:len(s.prev)-1]
}

func (s *state) push(newState int) {
	s.prev = append(s.prev, s.current)
	s.current = newState
}

func (s *state) replace(newState int) {
	s.current = newState
}

func (s *state) depth() int {
	return len(s.prev)
}

func (s *state) String() string {
	str := ""

	switch s.current {
	case start:
		str += "START"
	case objectStart:
		str += "OBJECTSTART"
	case object:
		str += "OBJECT"
	case arrayStart:
		str += "ARRAYSTART"
	case array:
		str += "ARRAY"
	case identifier:
		str += "IDENTIFIER"
	case postIdentifier:
		str += "POSTIDENTIFIER"
	case preValue:
		str += "PREVALUE"
	case stringValue:
		str += "VALUESTRING"
	case numberValue:
		str += "VALUENUMBER"
	case boolValue:
		str += "VALUEBOOL"
	case nullValue:
		str += "VALUENULL"

	default:
		str += fmt.Sprintf("UNKNOWN-%d", s.current)
	}

	// traverse s.prev in reverse order
	for i := len(s.prev) - 1; i >= 0; i-- {
		str += " < " + (&state{current: s.prev[i]}).String()
	}

	return str
}
