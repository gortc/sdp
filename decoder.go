package sdp

import (
	"fmt"
	"strconv"
	"time"

	"net"
	"unsafe"

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
	Version       int
	Origin        Origin
	Name          string
	Info          string
	Email         string
	Phone         string
	URI           string
	Connection    ConnectionData
	Attributes    Attributes
	Medias        Medias
	Encryption    Encryption
	Bandwidth     int
	BandwidthType BandwidthType
	Start         time.Time
	End           time.Time
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

// Encryption wraps encryption Key and Method.
type Encryption struct {
	Method string
	Key    string
}

// Media is media description and attributes.
type Media struct {
	Title         string
	Description   MediaDescription
	Connection    ConnectionData
	Attributes    Attributes
	Encryption    Encryption
	Bandwidth     int
	BandwidthType BandwidthType
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

func (d *Decoder) newFieldError(msg string) DecodeError {
	return DecodeError{
		Place:  fmt.Sprintf("%s/%s at line %d", d.section, d.t, d.pos),
		Reason: msg,
	}
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
	logger := log.WithField("t", t).WithFields(log.Fields{
		"s": s,
		"p": pos,
	})
	logger.Printf("isExpected(%s, %s, %d)", t, s, pos)
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
			logger.Printf("%s is expected", expected)
			return nil
		}
		if isOptional(expected) {
			continue
		}
		if isZeroOrMore(expected) {
			logger.Printf("%s is not necessary", expected)
			continue
		}
	}

	// checking possible section transitions
	switch s {
	case sectionSession:
		if pos < orderingAfterTime && isExpected(t, sectionTime, 0) == nil {
			logger.Printf("s->t")
			return nil
		}
		if isExpected(t, sectionMedia, 0) == nil {
			logger.Printf("s->m")
			return nil
		}
	case sectionTime:
		if isExpected(t, sectionSession, orderingAfterTime) == nil {
			logger.Printf("t->s")
			return nil
		}
		if isExpected(t, sectionMedia, 0) == nil {
			logger.Printf("t->m")
			return nil
		}
	case sectionMedia:
		if pos != 0 && isExpected(t, sectionMedia, 0) == nil {
			logger.Printf("m->m")
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

func (d *Decoder) decodeKV() (string, string) {
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
	return string(key), string(value)
}

func (d *Decoder) decodeTiming(m *Message) error {
	log.Println("decoding timing")
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
			d.pos--
			return nil
		}
	}
	return nil
}

func (d *Decoder) decodeMedia(m *Message) error {
	log.Println("decoding media")
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

func addAttribute(a Attributes, k, v string) Attributes {
	if a == nil {
		a = make(Attributes)
	}
	if len(v) == 0 {
		v = "true"
	}
	a[k] = v
	return a
}

func (d *Decoder) decodeAttribute(m *Message) error {
	k, v := d.decodeKV()
	switch d.section {
	case sectionMedia:
		d.m.Attributes = addAttribute(d.m.Attributes, k, v)
	default:
		m.Attributes = addAttribute(m.Attributes, k, v)
	}
	return nil
}

func (d *Decoder) decodeSessionName(m *Message) error {
	m.Name = string(d.v)
	return nil
}

func (d *Decoder) decodeSessionInfo(m *Message) error {
	if d.section == sectionMedia {
		d.m.Title = string(d.v)
	} else {
		m.Info = string(d.v)
	}
	return nil
}

func (d *Decoder) decodeEmail(m *Message) error {
	m.Email = string(d.v)
	return nil
}

func (d *Decoder) decodePhone(m *Message) error {
	m.Phone = string(d.v)
	return nil
}

func (d *Decoder) decodeURI(m *Message) error {
	m.URI = string(d.v)
	return nil
}

func (d *Decoder) decodeEncryption(m *Message) error {
	k, v := d.decodeKV()
	e := Encryption{
		Key:    v,
		Method: k,
	}
	switch d.section {
	case sectionMedia:
		d.m.Encryption = e
	default:
		m.Encryption = e
	}
	return nil
}

func decodeIP(dst net.IP, v []byte) (net.IP, error) {
	// ALLOCATIONS: suboptimal.
	return net.ParseIP(string(v)), nil
}

func decodeByte(dst []byte) (byte, error) {
	// ALLOCATIONS: suboptimal.
	n, err := strconv.ParseInt(string(dst), 10, 16)
	if err != nil {
		return 0, err
	}
	return byte(n), err
}

func (d *Decoder) decodeConnectionData(m *Message) error {
	// c=<nettype> <addrtype> <connection-address>
	var (
		netType           []byte
		addressType       []byte
		connectionAddress []byte
		subField          int
		err               error
	)
	for _, v := range d.v {
		if v == ' ' {
			subField++
			continue
		}
		switch subField {
		case 0:
			netType = append(netType, v)
		case 1:
			addressType = append(addressType, v)
		case 2:
			connectionAddress = append(connectionAddress, v)
		default:
			err = d.newFieldError("unexpected subfield count")
			return errors.Wrap(err, "failed to decode connection data")
		}
	}
	if len(netType) == 0 {
		err = d.newFieldError("nettype is empty")
		return errors.Wrap(err, "failed to decode connection data")
	}
	if len(addressType) == 0 {
		err = d.newFieldError("addrtype is empty")
		return errors.Wrap(err, "failed to decode connection data")
	}
	if len(connectionAddress) == 0 {
		err := d.newFieldError("connection-address is empty")
		return errors.Wrap(err, "failed to decode connection data")
	}
	m.Connection.AddressType = string(addressType)
	m.Connection.NetworkType = string(netType)

	// decoding address
	// <base multicast address>[/<ttl>]/<number of addresses>
	var (
		base   []byte
		first  []byte
		second []byte
	)
	subField = 0
	for _, v := range connectionAddress {
		if v == '/' {
			subField++
			continue
		}
		switch subField {
		case 0:
			base = append(base, v)
		case 1:
			first = append(first, v)
		case 3:
			second = append(second, v)
		default:
			err = d.newFieldError("unexpected fourth element in address")
			return errors.Wrap(err, "failed to decode connection data")
		}
	}
	m.Connection.IP, err = decodeIP(m.Connection.IP, base)
	if err != nil {
		return errors.Wrap(err, "failed to decode connection data")
	}
	if len(second) > 0 {
		m.Connection.TTL, err = decodeByte(first)
		if err != nil {
			return errors.Wrap(err, "failed to decode connection data")
		}
		m.Connection.Addresses, err = decodeByte(second)
		if err != nil {
			return errors.Wrap(err, "failed to decode connection data")
		}
	} else {
		m.Connection.Addresses, err = decodeByte(first)
		if err != nil {
			return errors.Wrap(err, "failed to decode connection data")
		}
	}
	return nil
}

func (d *Decoder) decodeBandwidth(m *Message) error {
	k, v := d.decodeKV()
	if len(v) == 0 {
		msg := "no value specified"
		err := newSectionDecodeError(d.section, msg)
		return errors.Wrap(err, "failed to decode bandwidth")
	}
	var (
		t   BandwidthType
		n   int
		err error
	)
	switch bandWidthType := BandwidthType(k); bandWidthType {
	case BandwidthApplicationSpecific, BandwidthConferenceTotal:
		t = bandWidthType
	default:
		msg := fmt.Sprintf("bad bandwidth type %s", k)
		err = newSectionDecodeError(d.section, msg)
		return errors.Wrap(err, "failed to decode bandwidth")
	}
	if n, err = strconv.Atoi(v); err != nil {
		return errors.Wrap(err, "failed to convert decode bandwidth")
	}
	if d.section == sectionMedia {
		d.m.BandwidthType = t
		d.m.Bandwidth = n
	} else {
		m.BandwidthType = t
		m.Bandwidth = n
	}
	return nil
}

func parseNTP(v []byte) (uint64, error) {
	return strconv.ParseUint(string(v), 10, 64)
}

func (d *Decoder) decodeTimingField(m *Message) error {
	var (
		startV, endV []byte
		isEndV       bool
		err          error
	)
	for _, v := range d.v {
		if v == ' ' {
			if isEndV {
				msg := "unexpected second space in timing"
				err = newSectionDecodeError(d.section, msg)
				return errors.Wrap(err, "failed to decode timing")
			}
			isEndV = true
			continue
		}
		if isEndV {
			endV = append(endV, v)
		} else {
			startV = append(startV, v)
		}
	}
	var (
		ntpStart, ntpEnd uint64
	)
	if ntpStart, err = parseNTP(startV); err != nil {
		return errors.Wrap(err, "failed to parse start time")
	}
	if ntpEnd, err = parseNTP(endV); err != nil {
		return errors.Wrap(err, "failed to parse end time")
	}
	m.Start = NTPToTime(ntpStart)
	m.End = NTPToTime(ntpEnd)
	return nil
}

func (d *Decoder) decodeField(m *Message) error {
	switch d.t {
	case TypeProtocolVersion:
		return d.decodeVersion(m)
	case TypeAttribute:
		return d.decodeAttribute(m)
	case TypeSessionName:
		return d.decodeSessionName(m)
	case TypeSessionInformation:
		return d.decodeSessionInfo(m)
	case TypeEmail:
		return d.decodeEmail(m)
	case TypePhone:
		return d.decodePhone(m)
	case TypeURI:
		return d.decodeURI(m)
	case TypeEncryptionKey:
		return d.decodeEncryption(m)
	case TypeBandwidth:
		return d.decodeBandwidth(m)
	case TypeTiming:
		return d.decodeTimingField(m)
	case TypeConnectionData:
		return d.decodeConnectionData(m)
	}
	// TODO: uncomment when all decoder methods implemented
	// panic("unexpected field")
	//log.Warnln("skipping decoding of", d.t)
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
			oldPosition := d.sPos
			if err := d.decodeTiming(m); err != nil {
				return errors.Wrap(err, "failed to decode timing")
			}
			d.sPos = oldPosition
			d.section = sectionSession
		case TypeMediaDescription:
			d.pos--
			oldPosition := d.sPos
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

// b2s converts byte slice to a string without memory allocation.
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
