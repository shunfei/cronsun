package utils

import (
	"errors"
)

type fmsState int

const (
	stateArgumentOutside fmsState = iota
	stateArgumentStart
	stateArgumentEnd
)

var errEndOfLine = errors.New("End of line")

type cmdArgumentParser struct {
	s            string
	i            int
	length       int
	state        fmsState
	startToken   byte
	shouldEscape bool
	currArgument []byte
	err          error
}

func newCmdArgumentParser(s string) *cmdArgumentParser {
	return &cmdArgumentParser{
		s:            s,
		i:            -1,
		length:       len(s),
		currArgument: make([]byte, 0, 16),
	}
}

func (cap *cmdArgumentParser) parse() (arguments []string) {
	for {
		cap.next()

		if cap.err != nil {
			if cap.shouldEscape {
				cap.currArgument = append(cap.currArgument, '\\')
			}

			if len(cap.currArgument) > 0 {
				arguments = append(arguments, string(cap.currArgument))
			}

			return
		}

		switch cap.state {
		case stateArgumentOutside:
			cap.detectStartToken()
		case stateArgumentStart:
			if !cap.detectEnd() {
				cap.detectContent()
			}
		case stateArgumentEnd:
			cap.state = stateArgumentOutside
			arguments = append(arguments, string(cap.currArgument))
			cap.currArgument = cap.currArgument[:0]
		}
	}
}

func (cap *cmdArgumentParser) previous() {
	if cap.i >= 0 {
		cap.i--
	}
}

func (cap *cmdArgumentParser) next() {
	if cap.length-cap.i == 1 {
		cap.err = errEndOfLine
		return
	}
	cap.i++
}

func (cap *cmdArgumentParser) detectStartToken() {
	c := cap.s[cap.i]
	if c == ' ' {
		return
	}

	switch c {
	case '\\':
		cap.startToken = 0
		cap.shouldEscape = true
	case '"', '\'':
		cap.startToken = c
	default:
		cap.startToken = 0
		cap.previous()
	}
	cap.state = stateArgumentStart
}

func (cap *cmdArgumentParser) detectContent() {
	c := cap.s[cap.i]

	if cap.shouldEscape {
		switch c {
		case ' ', '\\', cap.startToken:
			cap.currArgument = append(cap.currArgument, c)
		default:
			cap.currArgument = append(cap.currArgument, '\\', c)
		}
		cap.shouldEscape = false
		return
	}

	if c == '\\' {
		cap.shouldEscape = true
	} else {
		cap.currArgument = append(cap.currArgument, c)
	}
}

func (cap *cmdArgumentParser) detectEnd() (detected bool) {
	c := cap.s[cap.i]

	if cap.startToken == 0 {
		if c == ' ' && !cap.shouldEscape {
			cap.state = stateArgumentEnd
			cap.previous()
			return true
		}
		return false
	}

	if c == cap.startToken && !cap.shouldEscape {
		cap.state = stateArgumentEnd
		return true
	}

	return false
}

func ParseCmdArguments(s string) (arguments []string) {
	return newCmdArgumentParser(s).parse()
}
