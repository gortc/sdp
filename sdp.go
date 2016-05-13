// Package sdp implements RFC 4566 SDP: Session Description Protocol.
package sdp

import (
	"bytes"
	"errors"
	"fmt"
	"unicode/utf8"
)

const (
	lineDelimiter = '='
	newLine       = '\n'
)

// Line of SDP session.
//
// Form
// 	<type>=<value>
//
// Where <type> MUST be exactly one case-significant character and
// <value> is structured text whose format depends on <type>.
type Line struct {
	Type  Type
	Value []byte
}

func byteSliceEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if b[i] != a[i] {
			return false
		}
	}
	return true
}

// Equal returns true if l == b.
func (l Line) Equal(b Line) bool {
	if l.Type != b.Type {
		return false
	}
	return byteSliceEqual(l.Value, b.Value)
}

// Decode parses b into l and returns error if any.
//
// Decode does not reuse b, so it is safe to corrupt it.
func (l *Line) Decode(b []byte) error {
	delimiter := bytes.IndexRune(b, lineDelimiter)
	if delimiter == -1 {
		// TODO: replace with appropriate error
		return errors.New("no delim")
	}
	if len(b) < (delimiter + 1) {
		// TODO: replace with appropriate error
		return errors.New("too small")
	}
	r, _ := utf8.DecodeRune(b[:delimiter])
	l.Type = Type(r)
	l.Value = append(l.Value, b[delimiter+1:]...)
	return nil
}

func (l Line) String() string {
	return fmt.Sprintf("%s: %s",
		l.Type, string(l.Value),
	)
}

func appendRune(b []byte, r rune) []byte {
	buf := make([]byte, 4)
	n := utf8.EncodeRune(buf, r)
	b = append(b, buf[:n]...)
	return b
}

// AppendTo appends Line encoded value to b.
func (l Line) AppendTo(b []byte) []byte {
	b = l.Type.appendTo(b)
	b = appendRune(b, lineDelimiter)
	return append(b, l.Value...)
}

// Type of SDP Line is exactly one case-significant character.
type Type rune

func (t Type) appendTo(b []byte) []byte {
	return appendRune(b, rune(t))
}

func (t Type) String() string {
	switch t {
	case TypeAttributes:
		return "attributes"
	case TypePhone:
		return "phone"
	case TypeEmail:
		return "email"
	case TypeConnectionData:
		return "connection data"
	case TypeURI:
		return "uri"
	case TypeSessionName:
		return "session name"
	case TypeOrigin:
		return "origin"
	case TypeProtocolVersion:
		return "version"
	case TypeTiming:
		return "timing"
	case TypeBandwidth:
		return "bandwidth"
	case TypeSessionInformation:
		return "session info"
	case TypeRepeatTimes:
		return "repeat times"
	case TypeTimeZones:
		return "time zones"
	case TypeEncryptionKeys:
		return "encryption keys"
	case TypeMediaDescriptions:
		return "media descriptions"
	default:
		// falling back to raw value.
		return string(rune(t))
	}
}

// Attribute types as described in RFC 4566.
const (
	TypeProtocolVersion    Type = 'v'
	TypeOrigin             Type = 'o'
	TypeSessionName        Type = 's'
	TypeSessionInformation Type = 'i'
	TypeURI                Type = 'u'
	TypeEmail              Type = 'e'
	TypePhone              Type = 'p'
	TypeConnectionData     Type = 'c'
	TypeBandwidth          Type = 'b'
	TypeTiming             Type = 't'
	TypeRepeatTimes        Type = 'r'
	TypeTimeZones          Type = 'z'
	TypeEncryptionKeys     Type = 'k'
	TypeAttributes         Type = 'a'
	TypeMediaDescriptions  Type = 'm'
)

// Session is set of Lines.
type Session []Line

// AppendTo appends all session lines to b and returns b.
func (s Session) AppendTo(b []byte) []byte {
	last := len(s) - 1
	for i, l := range s {
		b = l.AppendTo(b)
		if i < last {
			// not adding newline on end
			b = appendRune(b, newLine)
		}
	}
	return b
}

func blankSlice(v []byte) bool {
	if v == nil {
		return true
	}
	if len(v) == 0 {
		return true
	}
	return false
}

// sliceScanner is custom in-memory scanner for slice
// that will scan all non-whitespace lines.
type sliceScanner struct {
	pos  int
	end  int
	v    []byte
	line []byte
}

func newScanner(v []byte) sliceScanner {
	return sliceScanner{
		v: v,
	}
}

func (s sliceScanner) Line() []byte {
	return s.line
}

func (s *sliceScanner) Scan() bool {
	// CPU: optimizations are possible.
	// TODO: handle /r
	for {
		s.pos = s.end
		if s.pos >= len(s.v) {
			// EOF
			s.line = s.line[:0]
			s.v = s.v[:0]
			return false
		}
		newLinePos := bytes.IndexRune(s.v[s.pos:], newLine)
		s.end = s.pos + newLinePos + 1
		if newLinePos < 0 {
			// next line symbol not found
			s.end = len(s.v)
		}
		s.line = bytes.TrimSpace(s.v[s.pos:s.end])
		if blankSlice(s.line) {
			continue
		}
		return true
	}
}

// DecodeSession decodes Session from b, returning error if any. Blank
// lines and leading/trialing whitespace are ignored.
//
// If s is passed, it will be reused with its lines.
// It is safe to corrupt b.
func DecodeSession(b []byte, s Session) (Session, error) {
	var (
		line Line
		err  error
	)
	scanner := newScanner(b)
	for scanner.Scan() {
		// trying to reuse some memory
		l := len(s)
		if cap(s) > l+1 {
			// picking element from s that is not in
			// slice bounds, but in underlying array
			// and reusing it byte slice
			line.Value = s[:l+1][l].Value[:0]
		}
		if err = line.Decode(scanner.Line()); err != nil {
			break
		}
		s = append(s, line)
		line.Value = nil // not corrupting.
	}
	return s, err
}
