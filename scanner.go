package jsonparser

import (
	"io"
)

const (
	ObjectStart = '{' // {
	ObjectEnd   = '}' // }
	String      = '"' // "
	Colon       = ':' // :
	Comma       = ',' // ,
	ArrayStart  = '[' // [
	ArrayEnd    = ']' // ]
	True        = 't' // t
	False       = 'f' // f
	Null        = 'n' // n
)

// NewScanner returns a new Scanner for the io.Reader r.
// A Scanner reads from the supplied io.Reader and produces via Next a stream
// of tokens, expressed as []byte slices.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		br: byteReader{
			r: r,
		},
	}
}

// Scanner implements a JSON scanner as defined in RFC 7159.
type Scanner struct {
	br     byteReader
	offset int
}

var whitespace = [256]bool{
	' ': true,
	'\r': true,
	'\n': true,
	'\t': true,
}

func (s *Scanner) Next() []byte {
	s.br.release(s.offset)
	w := s.br.window()

	for {
		for pos, c := range w {
			if whitespace[c] {
				continue
			}


			switch c {
			case ObjectStart, ObjectEnd, Colon, Comma, ArrayStart, ArrayEnd:
				s.offset = pos + 1
				return w[pos:s.offset]
			}

			s.br.release(pos)

			switch c {
			case True:
				s.offset = s.validateToken("true")
			case False:
				s.offset = s.validateToken("false")
			case Null:
				s.offset = s.validateToken("null")
			case String:
				if s.parseString() < 2 {
					return nil
				}
			default:
				s.offset = s.parseNumber(c)
			}
			return s.br.window()[:s.offset]
		}

		// it's all whitespace then, ignore it
		s.br.release(len(w))

		// refill buffer
		if s.br.extend() == 0 {
			// oef
			return nil
		}
		w = s.br.window()
	}
}

func (s *Scanner) validateToken(expected string) int {
	for {
		w := s.br.window()
		n := len(expected)

		if len(w) >= n {
			if string(w[:n]) != expected {
				return 0
			}
			return n
		}

		if s.br.extend() == 0 {
			return 0
		}
	}
}

// parseString returns the length of the string token
// located at the start of the window or 0 if there is no closing
// " before the end of the byteReader.
func (s *Scanner) parseString() int {
	escaped := false
	w := s.br.window()[1:]
	offset := 0

	for {
		for _, c := range w {
			switch {
			case c == '\\':
				escaped = true
			case c == '"' && !escaped:
				// finished
				s.offset = offset + 1
				return s.offset
			case escaped:
				escaped = false
			}
			offset++
		}

		// need more data from the pipe
		if s.br.extend() == 0 {
			// EOF.
			return 0
		}

		w = s.br.window()[offset:]
	}
}

func (s *Scanner) parseNumber(c byte) int {
	const (
		begin = iota
		leadingzero
		anydigit1
		decimal
		anydigit2
		exponent
		expsign
		anydigit3
	)

	offset := 0
	w := s.br.window()
	// int vs uint8 costs 10% on canada.json
	var state uint8 = begin
	if c == '-' {
		offset++
	}

	for {
		for _, elem := range w[offset:] {
			switch state {
			case begin:
				if elem >= '1' && elem <= '9' {
					state = anydigit1
				} else if elem == '0' {
					state = leadingzero
				} else {
					// error
					return 0
				}
			case anydigit1:
				if elem >= '0' && elem <= '9' {
					// stay in this state
					break
				}
				fallthrough
			case leadingzero:
				if elem == '.' {
					
				}
			}
		}
	}
}
