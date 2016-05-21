package sdp

import (
	"fmt"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

// Attributes is set of k:v.
type Attributes map[string]string

// Value returns string v.
func (a Attributes) Value(attribute string) string { return a[attribute] }

// Flag returns true if set.
func (a Attributes) Flag(flag string) bool { return len(a[flag]) > 0 }

// Message is top level abstraction.
type Message struct {
	Session Session

	Version    int
	Origin     Origin
	Name       string
	Info       string
	Email      string
	Phone      string
	Connection ConnectionData
	Attributes Attributes
	Medias     Medias
}

// Flag returns true if set.
func (m Message) Flag(flag string) bool {
	if len(m.Attributes) > 0 {
		return m.Attributes.Flag(flag)
	}
	return false
}

const blank = ""

// Attribute returns string v.
func (m Message) Attribute(attribute string) string {
	if len(m.Attributes) > 0 {
		m.Attributes.Value(attribute)
	}

	return blank
}

// Medias is list of Media.
type Medias []Media

// Media is media description and attributes.
type Media struct {
	Name        string
	Info        string
	Description MediaDescription
	Connection  ConnectionData
	Attributes  Attributes
}

// Decoder decodes session.
type Decoder struct {
	s       Session
	pos     int
	t       Type
	v       []byte
	l       Line
	section section
	sPos    int
	m       Media
}

func (d *Decoder) next() bool {
	//time.Sleep(time.Millisecond * 100)
	if d.pos >= len(d.s) {
		return false
	}
	d.l = d.s[d.pos]
	d.v = d.l.Value
	d.t = d.l.Type
	d.pos++
	return true
}

func (d *Decoder) line() Line {
	return d.l
}

type section int

const (
	sectionSession section = iota
	sectionTime
	sectionMedia
)

func (s section) String() string {
	switch s {
	case sectionSession:
		return "s"
	case sectionTime:
		return "t"
	case sectionMedia:
		return "m"
	default:
		panic("unexpected")
	}
}

type ordering []Type

var orderingSession = ordering{
	TypeProtocolVersion,
	TypeOrigin,
	TypeSessionName,
	TypeSessionInformation,
	TypeURI,
	TypeEmail,
	TypePhone,
	TypeConnectionData,
	TypeBandwidth, // 0 or more
	TypeTimeZones, // ordering after time start
	TypeEncryptionKey,
	TypeAttribute, // 0 or more
}

const orderingAfterTime = 9

var orderingTime = ordering{
	TypeTiming,
	TypeRepeatTimes,
}

var orderingMedia = ordering{
	TypeMediaDescription,
	TypeSessionInformation, // title
	TypeConnectionData,
	TypeBandwidth,
	TypeEncryptionKey,
	TypeAttribute,
}

// isExpected determines if t is expected on pos in s section and returns nil,
// if it is expected and DecodeError if not.
func isExpected(t Type, s section, pos int) error {
	//logger := log.WithField("t", t).WithFields(log.Fields{
	//	"s": s,
	//	"p": pos,
	//})
	//logger.Printf("isExpected(%s, %s, %d)", t, s, pos)
	o := getOrdering(s)
	if len(o) <= pos {
		msg := fmt.Sprintf("position %d is out of range (>%d)",
			pos, len(o),
		)
		err := newSectionDecodeError(s, msg)
		return errors.Wrapf(err, "field %s is unexpected", t)
	}
	for _, expected := range o[pos:] {
		if expected == t {
			//logger.Printf("%s is expected", expected)
			return nil
		}
		if isOptional(expected) {
			continue
		}
		if isZeroOrMore(expected) {
			//logger.Printf("%s is not necessary", expected)
			continue
		}
	}

	// checking possible section transitions
	switch s {
	case sectionSession:
		if pos < orderingAfterTime && isExpected(t, sectionTime, 0) == nil {
			//logger.Printf("s->t")
			return nil
		}
		if isExpected(t, sectionMedia, 0) == nil {
			//logger.Printf("s->m")
			return nil
		}
	case sectionTime:
		if isExpected(t, sectionSession, orderingAfterTime) == nil {
			//logger.Printf("t->s")
			return nil
		}
		if isExpected(t, sectionMedia, 0) == nil {
			//logger.Printf("t->m")
			return nil
		}
	case sectionMedia:
		if isExpected(t, sectionMedia, 0) == nil {
			//logger.Printf("m->m")
			return nil
		}
	}
	msg := fmt.Sprintf("no matches in ordering array at %s[%d]",
		s, pos,
	)
	err := newSectionDecodeError(s, msg)
	return errors.Wrapf(err, "field %s is unexpected", t)
}

func getOrdering(s section) ordering {
	switch s {
	case sectionSession:
		return orderingSession
	case sectionMedia:
		return orderingMedia
	case sectionTime:
		return orderingTime
	default:
		panic("unexpected section")
	}
}

func isOptional(t Type) bool {
	switch t {
	case TypeProtocolVersion, TypeOrigin, TypeSessionName:
		return false
	case TypeTiming:
		return false
	case TypeMediaDescription:
		return false
	default:
		return true
	}
}

func isZeroOrMore(t Type) bool {
	switch t {
	case TypeBandwidth, TypeAttribute:
		return true
	default:
		return false
	}
}

func newSectionDecodeError(s section, m string) DecodeError {
	place := fmt.Sprintf("section %s", s)
	return newDecodeError(place, m)
}

func (d *Decoder) decodeTiming(m *Message) error {
	//log.Println("decoding timing")
	d.sPos = 0
	d.section = sectionTime
	for d.next() {
		if err := isExpected(d.t, d.section, d.sPos); err != nil {
			return errors.Wrap(err, "decode failed")
		}
		if !isZeroOrMore(d.t) {
			d.sPos++
		}
		switch d.t {
		case TypeTiming, TypeRepeatTimes:
			return d.decodeField(m)
		default:
			// possible switch to Media or Session description
			//log.Warn("rewinding, out of timing")
			d.pos--
			return nil
		}
	}
	return nil
}

func (d *Decoder) decodeMedia(m *Message) error {
	d.sPos = 0
	d.section = sectionMedia
	d.m = Media{}
	for d.next() {
		if err := isExpected(d.t, d.section, d.sPos); err != nil {
			return errors.Wrap(err, "decode failed")
		}
		if d.t == TypeMediaDescription && d.sPos != 0 {
			d.pos--
			break
		}
		if !isZeroOrMore(d.t) {
			d.sPos++
		}
		if err := d.decodeField(m); err != nil {
			return errors.Wrap(err, "failed to decode field")
		}
	}
	m.Medias = append(m.Medias, d.m)
	return nil
}

func (d *Decoder) decodeVersion(m *Message) error {
	n, err := strconv.Atoi(string(d.v))
	if err != nil {
		return errors.Wrap(err, "failed to parse version")
	}
	m.Version = n
	return nil
}

func addAttribute(a Attributes, k, v []byte) Attributes {
	if a == nil {
		a = make(Attributes)
	}
	if len(v) == 0 {
		v = append(v, "true"...)
	}
	a[string(k)] = string(v)
	return a
}

func (d *Decoder) decodeAttribute(m *Message) error {
	//log.Println("DECODING ATTRIBUTE")
	var (
		key     []byte
		value   []byte
		isValue bool
	)
	for _, v := range d.v {
		if v == ':' && !isValue {
			isValue = true
			continue
		}
		if isValue {
			value = append(value, v)
		} else {
			key = append(key, v)
		}
	}
	switch d.section {
	case sectionMedia:
		d.m.Attributes = addAttribute(d.m.Attributes, key, value)
	default:
		m.Attributes = addAttribute(m.Attributes, key, value)
	}
	return nil
}

func (d *Decoder) decodeField(m *Message) error {
	switch d.t {
	case TypeProtocolVersion:
		return d.decodeVersion(m)
	case TypeAttribute:
		return d.decodeAttribute(m)
	}
	// TODO: uncomment when all decoder methods implemented
	// panic("unexpected field")
	log.Warnln("skipping decoding of", d.t)
	return nil
}

func (d *Decoder) decodeSession(m *Message) error {
	d.sPos = 0
	d.section = sectionSession
	for d.next() {
		if err := isExpected(d.t, d.section, d.sPos); err != nil {
			return errors.Wrap(err, "decode failed")
		}
		if !isZeroOrMore(d.t) {
			d.sPos++
		}
		switch d.t {
		case TypeTiming:
			d.pos--
			//log.Warn("rewinding, switching to T")
			if err := d.decodeTiming(m); err != nil {
				return errors.Wrap(err, "failed to decode timing")
			}
		case TypeMediaDescription:
			d.pos--
			oldPosition := d.sPos
			//log.Warn("rewinding, switching to M")
			if err := d.decodeMedia(m); err != nil {
				return errors.Wrap(err, "failed to decode media")
			}
			d.sPos = oldPosition
			d.section = sectionSession
		default:
			if err := d.decodeField(m); err != nil {
				return errors.Wrap(err, "failed to decode field")
			}
		}
	}
	return nil
}

// Decode message from session.
func (d *Decoder) Decode(m *Message) error {
	return d.decodeSession(m)
}
