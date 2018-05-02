package sdp

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/pkg/errors"
)

// Attributes is set of k:v.
type Attributes map[string][]string

// Value returns value of first attribute.
func (a Attributes) Value(attribute string) string {
	if len(a[attribute]) == 0 {
		return blank
	}
	return a[attribute][0]
}

// Values returns list of values associated to attribute.
func (a Attributes) Values(attribute string) []string {
	return a[attribute]
}

// Flag returns true if set.
func (a Attributes) Flag(flag string) bool {
	return len(a[flag]) != 0
}

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
	Bandwidths    map[BandwidthType]int
	BandwidthType BandwidthType
	Timing        []Timing
	TZAdjustments []TimeZone
}

// Timing wraps "repeat times" and 	"timing" information.
type Timing struct {
	Start   time.Time
	End     time.Time
	Repeat  time.Duration
	Active  time.Duration
	Offsets []time.Duration
}

// Start returns start of session.
func (m Message) Start() time.Time {
	if len(m.Timing) == 0 {
		return time.Time{}
	}
	return m.Timing[0].Start
}

// End returns end of session.
func (m Message) End() time.Time {
	if len(m.Timing) == 0 {
		return time.Time{}
	}
	return m.Timing[0].End
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
		return m.Attributes.Value(attribute)
	}
	return blank
}

// AddAttribute appends new k-v pair to attribute list.
func (m *Message) AddAttribute(k, v string) {
	m.Attributes = addAttribute(m.Attributes, k, v)
}

// AddFlag appends new flag to attribute list.
func (m *Message) AddFlag(f string) {
	m.AddAttribute(f, blank)
}

// Medias is list of Media.
type Medias []Media

// Encryption wraps encryption Key and Method.
type Encryption struct {
	Method string
	Key    string
}

// Blank determines whether Encryption is blank value.
func (e Encryption) Blank() bool {
	return e.Equal(Encryption{})
}

// Equal returns e == b.
func (e Encryption) Equal(b Encryption) bool {
	return e == b
}

// Bandwidths is map of BandwidthsType and int (bytes per second).
type Bandwidths map[BandwidthType]int

// Media is media description and attributes.
type Media struct {
	Title       string
	Description MediaDescription
	Connection  ConnectionData
	Attributes  Attributes
	Encryption  Encryption
	Bandwidths  Bandwidths
}

// AddAttribute appends new k-v pair to attribute list.
func (m *Media) AddAttribute(k string, values ...string) {
	v := strings.Join(values, " ")
	m.Attributes = addAttribute(m.Attributes, k, v)
}

// AddFlag appends new flag to attribute list.
func (m *Media) AddFlag(f string) {
	m.AddAttribute(f, blank)
}

// Flag returns true if set.
func (m *Media) Flag(f string) bool {
	return m.Attributes.Flag(f)
}

// Attribute returns string v.
func (m *Media) Attribute(k string) string {
	return m.Attributes.Value(k)
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

// NewDecoder returns Decoder for Session.
func NewDecoder(s Session) Decoder {
	return Decoder{
		s: s,
	}
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
	TypeBandwidth,     // 0 or more
	TypeTimeZones,     // *
	TypeEncryptionKey, // ordering after time start
	TypeAttribute,     // 0 or more
}

const orderingAfterTime = 10

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
	if len(o) > pos {
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
		if pos != 0 && isExpected(t, sectionMedia, 0) == nil {
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
			if err := d.decodeField(m); err != nil {
				return errors.Wrap(err, "decode failed")
			}
		default:
			// possible switch to Media or Session description
			d.pos--
			return nil
		}
	}
	return nil
}

func (d *Decoder) decodeMedia(m *Message) error {
	//log.Println("decoding media")
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
		v = blank
	}
	a[k] = append(a[k], v)
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

func isIPv4(ip net.IP) bool {
	return ip.To4() != nil
}

func (d *Decoder) decodeConnectionData(m *Message) error {
	// TODO: simplify.
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
	switch d.section {
	case sectionMedia:
		d.m.Connection.AddressType = string(addressType)
		d.m.Connection.NetworkType = string(netType)
	case sectionSession:
		m.Connection.AddressType = string(addressType)
		m.Connection.NetworkType = string(netType)
	}
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
		case 2:
			second = append(second, v)
		default:
			err = d.newFieldError("unexpected fourth element in address")
			return errors.Wrap(err, "failed to decode connection data")
		}
	}
	switch d.section {
	case sectionMedia:
		d.m.Connection.IP, err = decodeIP(m.Connection.IP, base)
	case sectionSession:
		m.Connection.IP, err = decodeIP(m.Connection.IP, base)
	}
	if err != nil {
		return errors.Wrap(err, "failed to decode connection data")
	}
	isV4 := isIPv4(m.Connection.IP)
	if len(second) > 0 {
		if !isV4 {
			err := d.newFieldError("unexpected TTL for IPv6")
			return errors.Wrap(err, "failed to decode connection data")
		}
		switch d.section {
		case sectionMedia:
			d.m.Connection.TTL, err = decodeByte(first)
		case sectionSession:
			m.Connection.TTL, err = decodeByte(first)
		}
		if err != nil {
			return errors.Wrap(err, "failed to decode connection data")
		}
		switch d.section {
		case sectionMedia:
			d.m.Connection.Addresses, err = decodeByte(second)
		case sectionSession:
			m.Connection.Addresses, err = decodeByte(second)
		}
		if err != nil {
			return errors.Wrap(err, "failed to decode connection data")
		}
	} else if len(first) > 0 {
		if isV4 {
			switch d.section {
			case sectionMedia:
				m.Connection.TTL, err = decodeByte(first)
			case sectionSession:
				m.Connection.TTL, err = decodeByte(first)
			}
		} else {
			switch d.section {
			case sectionMedia:
				d.m.Connection.Addresses, err = decodeByte(second)
			case sectionSession:
				m.Connection.Addresses, err = decodeByte(second)
			}
		}
		if err != nil {
			msg := fmt.Sprintf("bad connection data <%s> at <%s>",
				b2s(first), b2s(connectionAddress),
			)
			return errors.Wrap(err, msg)
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
		if d.m.Bandwidths == nil {
			d.m.Bandwidths = make(Bandwidths)
		}
		d.m.Bandwidths[t] = n
	} else {
		if m.Bandwidths == nil {
			m.Bandwidths = make(Bandwidths)
		}
		m.Bandwidths[t] = n
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
	t := Timing{}
	t.Start = NTPToTime(ntpStart)
	t.End = NTPToTime(ntpEnd)
	m.Timing = append(m.Timing, t)
	return nil
}

const (
	fieldsDelimiter = ' '
)

func decodeString(v []byte, s *string) error {
	*s = b2s(v)
	return nil
}

func decodeInt(v []byte, i *int) error {
	var err error
	*i, err = strconv.Atoi(b2s(v))
	return err
}

func subfields(v []byte) [][]byte {
	return bytes.Split(v, []byte{fieldsDelimiter})
}

func (d *Decoder) subfields() [][]byte {
	return subfields(d.v)
}

func (d *Decoder) decodeOrigin(m *Message) error {
	// o=0<username> 1<sess-id> 2<sess-version> 3<nettype> 4<addrtype>
	// 5<unicast-address>
	// ALLOCATIONS: suboptimal
	// CPU: suboptimal
	var (
		err error
	)
	p := d.subfields()
	if len(p) != 6 {
		msg := fmt.Sprintf("unexpected subfields count %d != %d", len(p), 6)
		err = newSectionDecodeError(d.section, msg)
		return errors.Wrap(err, "failed to decode origin")
	}
	o := m.Origin
	if err = decodeString(p[0], &o.Username); err != nil {
		return errors.Wrap(err, "failed to decode username")
	}
	if err = decodeInt(p[1], &o.SessionID); err != nil {
		return errors.Wrap(err, "failed to decode sess-id")
	}
	if err = decodeInt(p[2], &o.SessionVersion); err != nil {
		return errors.Wrap(err, "failed to decode sess-version")
	}
	if err = decodeString(p[3], &o.NetworkType); err != nil {
		return errors.Wrap(err, "failed to decode net-type")
	}
	if err = decodeString(p[4], &o.AddressType); err != nil {
		return errors.Wrap(err, "failed to decode addres-type")
	}
	if err = decodeString(p[5], &o.Address); err != nil {
		return errors.Wrap(err, "failed to decode address")
	}
	m.Origin = o
	return nil
}

func decodeInterval(b []byte, v *time.Duration) error {
	if len(b) == 1 && b[0] == '0' {
		*v = 0
		return nil
	}
	var (
		unit            time.Duration
		noUnitSpecified bool
		val             int
	)
	switch b[len(b)-1] {
	case 'd':
		unit = time.Hour * 24
	case 'h':
		unit = time.Hour
	case 'm':
		unit = time.Minute
	case 's':
		unit = time.Second
	default:
		unit = time.Second
		noUnitSpecified = true
	}
	if !noUnitSpecified {
		if len(b) < 2 {
			err := io.ErrUnexpectedEOF
			return errors.Wrap(err, "unit without value is invalid duration")
		}
		b = b[:len(b)-1]
	}
	if err := decodeInt(b, &val); err != nil {
		return errors.Wrap(err, "unable to decode value")
	}
	*v = time.Duration(val) * unit
	return nil
}

func shouldBePositive(i int) {
	if i <= 0 {
		panic("value should be positive")
	}
}

func (d *Decoder) decodeRepeatTimes(m *Message) error {
	// r=0<repeat interval> 1<active duration> 2<offsets from start-time>
	shouldBePositive(len(m.Timing)) // should be newer blank
	p := d.subfields()
	var err error
	if len(p) < 3 {
		msg := fmt.Sprintf("unexpected subfields count %d < 3", len(p))
		err = newSectionDecodeError(d.section, msg)
		return errors.Wrap(err, "failed to decode repeat")
	}
	t := m.Timing[len(m.Timing)-1]
	if err = decodeInterval(p[0], &t.Repeat); err != nil {
		return errors.Wrap(err, "failed to decode repeat interval")
	}
	if err = decodeInterval(p[1], &t.Active); err != nil {
		return errors.Wrap(err, "failed to decode active duration")
	}
	var dd time.Duration
	for i, pp := range p[2:] {
		if err = decodeInterval(pp, &dd); err != nil {
			return errors.Wrapf(err, "failed to decode offset %d", i)
		}
		t.Offsets = append(t.Offsets, dd)
	}
	return nil
}

func (d *Decoder) decodeTimeZoneAdjustments(m *Message) error {
	// z=<adjustment time> <offset> <adjustment time> <offset> ....
	p := d.subfields()
	var (
		adjustment TimeZone
		t          uint64
		err        error
	)
	if len(p)%2 != 0 {
		msg := fmt.Sprintf("unexpected subfields count %d", len(p))
		err = newSectionDecodeError(d.section, msg)
		return errors.Wrap(err, "failed to decode tz-adjustments")
	}
	for i := 0; i < len(p); i += 2 {
		if t, err = parseNTP(p[i]); err != nil {
			return errors.Wrap(err, "failed to decode adjustment start")
		}
		adjustment.Start = NTPToTime(t)
		if err = decodeInterval(p[i+1], &adjustment.Offset); err != nil {
			return errors.Wrap(err, "failed to decode offset")
		}
		m.TZAdjustments = append(m.TZAdjustments, adjustment)
	}
	return nil
}

func (d *Decoder) decodeMediaDescription(m *Message) error {
	// m=0<media> 1<port> 2<proto> 3<fmt> ...
	var (
		desc MediaDescription
		err  error
	)
	p := d.subfields()
	if len(p) < 4 {
		msg := fmt.Sprintf("unexpected subfields count %d < 4", len(p))
		err = newSectionDecodeError(d.section, msg)
		return errors.Wrap(err, "failed to decode media description")
	}
	if err = decodeString(p[0], &desc.Type); err != nil {
		return errors.Wrap(err, "failed to decode media type")
	}
	// port: port/ports_number
	pp := bytes.Split(p[1], []byte{'/'})
	if err = decodeInt(pp[0], &desc.Port); err != nil {
		return errors.Wrap(err, "failed to decode port")
	}
	if len(pp) > 1 {
		if err = decodeInt(pp[1], &desc.PortsNumber); err != nil {
			return errors.Wrap(err, "failed to decode ports number")
		}
	}
	if err = decodeString(p[2], &desc.Protocol); err != nil {
		return errors.Wrap(err, "failed to decode protocol")
	}
	desc.Format = string(bytes.Join(p[3:], []byte{fieldsDelimiter}))
	d.m.Description = desc
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
	case TypeOrigin:
		return d.decodeOrigin(m)
	case TypeRepeatTimes:
		return d.decodeRepeatTimes(m)
	case TypeTimeZones:
		return d.decodeTimeZoneAdjustments(m)
	case TypeMediaDescription:
		return d.decodeMediaDescription(m)
	default:
		panic("unexpected field")
	}
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
