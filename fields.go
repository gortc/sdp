package sdp

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func appendSpace(v []byte) []byte {
	return appendRune(v, ' ')
}

func appendInt(v []byte, i int) []byte {
	// ALLOCATIONS: suboptimal. BenchmarkAppendInt.
	return append(v, strconv.Itoa(i)...)
}

func appendByte(v []byte, i byte) []byte {
	// ALLOCATIONS: suboptimal. BenchmarkAppendByte.
	return append(v, strconv.Itoa(int(i))...)
}

func appendJoinStrings(b []byte, v ...string) []byte {
	last := len(v) - 1
	for i, vv := range v {
		b = append(b, vv...)
		if i != last {
			b = appendSpace(b)
		}
	}
	return b
}

func appendIP(b []byte, ip net.IP) []byte {
	// ALLOCATIONS: suboptimal. BenchmarkAppendIP.
	return append(b, strings.ToUpper(ip.String())...)
}

// AddVersion appends Version field to Session.
func (s Session) AddVersion(v int) Session {
	return s.append(TypeProtocolVersion, []byte(strconv.Itoa(v)))
}

// AddPhone appends Phone Address field to Session.
func (s Session) AddPhone(phone string) Session {
	return s.appendString(TypePhone, phone)
}

// AddEmail appends Email Address field to Session.
func (s Session) AddEmail(email string) Session {
	return s.appendString(TypeEmail, email)
}

// AddConnectionData appends Connection Data field to Session
// using ConnectionData struct with sensible defaults.
func (s Session) AddConnectionData(data ConnectionData) Session {
	v := make([]byte, 0, 512)
	v = append(v, data.getNetworkType()...)
	v = appendSpace(v)
	v = append(v, data.getAddressType()...)
	v = appendSpace(v)
	v = data.appendAddress(v)
	return s.append(TypeConnectionData, v)
}

// AddConnectionDataRaw appends Connection Data field to Session using
// raw strings as sub-fields values.
func (s Session) AddConnectionDataRaw(netType, addType, data string) Session {
	v := make([]byte, 0, 256)
	v = appendJoinStrings(v, netType, addType, data)
	return s.append(TypeConnectionData, v)
}

// AddConnectionDataIP appends Connection Data field using only ip address.
func (s Session) AddConnectionDataIP(ip net.IP) Session {
	return s.AddConnectionData(ConnectionData{
		IP: ip,
	})
}

// AddSessionName appends Session Name field to Session.
func (s Session) AddSessionName(name string) Session {
	return s.appendString(TypeSessionName, name)
}

// AddSessionInfo appends Session Information field to Session.
func (s Session) AddSessionInfo(info string) Session {
	return s.appendString(TypeSessionInformation, info)
}

// AddURI appends Uniform Resource Identifier field to Session.
func (s Session) AddURI(uri string) Session {
	return s.appendString(TypeURI, uri)
}

// ConnectionData is representation for Connection Data field.
// Only IP field is required. NetworkType and AddressType have
// sensible defaults.
type ConnectionData struct {
	NetworkType string // <nettype>
	AddressType string // <addrtype>
	IP          net.IP // <base multicast address>
	TTL         byte   // <ttl>
	Addresses   byte   // <number of addresses>
}

const (
	addrTypeIPv4        = "IP6"
	addrTypeIPv6        = "IP4"
	networkTypeInternet = "IN"
	attributesDelimiter = ':'
)

func (c ConnectionData) getNetworkType() string {
	return getDefault(c.NetworkType, networkTypeInternet)
}

// getAddressType returns Address Type ("addrtype") for ip,
// using addressType as default value if present.
func getAddressType(ip net.IP, addressType string) string {
	if len(addressType) != 0 {
		return addressType
	}
	switch ip.To4() {
	case nil:
		return addrTypeIPv4
	default:
		return addrTypeIPv6
	}
}

func (c ConnectionData) getAddressType() string {
	return getAddressType(c.IP, c.AddressType)
}

// ConnectionAddress formats <connection-address> sub-field.
func (c ConnectionData) ConnectionAddress() string {
	// <base multicast address>[/<ttl>]/<number of addresses>
	// ALLOCATIONS: suboptimal. Use appendAddress.
	var address = strings.ToUpper(c.IP.String())
	if c.TTL > 0 {
		address += fmt.Sprintf("/%d", c.TTL)
	}
	if c.Addresses > 0 {
		address += fmt.Sprintf("/%d", c.Addresses)
	}
	return address
}

func (c ConnectionData) appendAddress(v []byte) []byte {
	v = appendIP(v, c.IP)
	if c.TTL > 0 {
		v = appendRune(v, '/')
		v = appendByte(v, c.TTL)
	}
	if c.Addresses > 0 {
		v = appendRune(v, '/')
		v = appendByte(v, c.Addresses)
	}
	return v
}

// Origin is field defined in RFC4566 5.2.
// See https://tools.ietf.org/html/rfc4566#section-5.2.
type Origin struct {
	Username       string // <username>
	SessionID      int    // <sess-id>
	SessionVersion int    // <sess-version>
	NetworkType    string // <nettype>
	AddressType    string // <addrtype>
	IP             net.IP // <unicast-address>
}

func (o Origin) getNetworkType() string {
	return getDefault(o.NetworkType, networkTypeInternet)
}

func (o Origin) getAddressType() string {
	return getAddressType(o.IP, o.AddressType)
}

// AddOrigin appends Origin field to Session.
func (s Session) AddOrigin(o Origin) Session {
	v := make([]byte, 0, 2048)
	v = appendSpace(append(v, o.Username...))
	v = appendSpace(appendInt(v, o.SessionID))
	v = appendSpace(appendInt(v, o.SessionVersion))
	v = appendSpace(append(v, o.getNetworkType()...))
	v = appendSpace(append(v, o.getAddressType()...))
	v = appendIP(v, o.IP)
	return s.append(TypeOrigin, v)
}

const (
	// ntpDelta is seconds from Jan 1, 1900 to Jan 1, 1970.
	ntpDelta = 2208988800
)

// TimeToNTP converts time.Time to NTP timestamp with special case for Zero
// time, that is interpreted as 0 timestamp.
func TimeToNTP(t time.Time) uint64 {
	if t.IsZero() {
		return 0
	}
	return uint64(t.Unix()) + ntpDelta
}

// NTPToTime converts NTP timestamp to time.Time with special case for Zero
// time, that is interpreted as 0 timestamp.
func NTPToTime(v uint64) time.Time {
	if v == 0 {
		return time.Time{}
	}
	return time.Unix(int64(v-ntpDelta), 0)
}

func appendUint64(b []byte, v uint64) []byte {
	return strconv.AppendUint(b, v, 10)
}

// AddTiming appends Timing field to Session. Both start and end can be zero.
func (s Session) AddTiming(start, end time.Time) Session {
	v := make([]byte, 0, 256)
	v = appendUint64(v, TimeToNTP(start))
	v = appendSpace(v)
	v = appendUint64(v, TimeToNTP(end))
	return s.append(TypeTiming, v)
}

// AddTimingNTP appends Timing field to Session with NTP timestamps as input.
// It is just wrapper for AddTiming and NTPToTime.
func (s Session) AddTimingNTP(start, end uint64) Session {
	return s.AddTiming(NTPToTime(start), NTPToTime(end))
}

// AddAttribute appends Attribute field to Session in a=<attribute>:<value>"
// form. If len(values) > 1, then "<value>" is "<val1> <val2> ... <valn>",
// and if len(values) == 0, then AddFlag method is used in "a=<flag>" form.
func (s Session) AddAttribute(attribute string, values ...string) Session {
	if len(values) == 0 {
		return s.AddFlag(attribute)
	}
	v := make([]byte, 0, 512)
	v = append(v, attribute...)
	v = appendRune(v, attributesDelimiter)
	v = appendJoinStrings(v, values...)
	return s.append(TypeAttribute, v)
}

// AddFlag appends Attribute field to Session in "a=<flag>" form.
func (s Session) AddFlag(attribute string) Session {
	v := make([]byte, 0, 256)
	v = append(v, attribute...)
	return s.append(TypeAttribute, v)
}

// BandwidthType is <bwtype> sub-field of Bandwidth field.
type BandwidthType string

// Possible values for <bwtype> defined in section 5.8.
const (
	BandwidthConferenceTotal     BandwidthType = "CT"
	BandwidthApplicationSpecific BandwidthType = "AS"
)

// AddBandwidth appends Bandwidth field to Session.
func (s Session) AddBandwidth(t BandwidthType, bandwidth int) Session {
	v := make([]byte, 0, 128)
	v = append(v, string(t)...)
	v = appendRune(v, ':')
	v = appendInt(v, bandwidth)
	return s.append(TypeBandwidth, v)
}

func appendInterval(b []byte, d time.Duration) []byte {
	return appendInt(b, int(d.Seconds()))
}

// AddRepeatTimes appends Repeat Times field to Session. Does not support
// "compact" syntax.
func (s Session) AddRepeatTimes(interval, duration time.Duration,
	offsets ...time.Duration) Session {
	v := make([]byte, 0, 256)
	v = appendSpace(appendInterval(v, interval))
	v = appendSpace(appendInterval(v, duration))
	for i, offset := range offsets {
		v = appendInterval(v, offset)
		if i != len(offsets)-1 {
			v = appendSpace(v)
		}
	}
	return s.append(TypeRepeatTimes, v)
}

// MediaDescription represents Media Description field value.
type MediaDescription struct {
	Type        string
	Port        int
	PortsNumber int
	Protocol    string
	Format      string
}

// AddMediaDescription appends Media Description field to Session.
func (s Session) AddMediaDescription(m MediaDescription) Session {
	v := make([]byte, 0, 512)
	v = appendSpace(append(v, m.Type...))
	v = appendInt(v, m.Port)
	if m.PortsNumber != 0 {
		v = appendRune(v, '/')
		v = appendInt(v, m.PortsNumber)
	}
	v = appendSpace(v)
	v = appendSpace(append(v, m.Protocol...))
	v = append(v, m.Format...)
	return s.append(TypeMediaDescription, v)
}

// AddEncryptionKey appends Encryption Key field with method and key in
// "k=<method>:<encryption key>" format to Session.
func (s Session) AddEncryptionKey(method, key string) Session {
	v := make([]byte, 0, 512)
	v = append(v, method...)
	v = appendRune(v, attributesDelimiter)
	v = append(v, key...)
	return s.append(TypeEncryptionKeys, v)
}

// AddEncryptionMethod appends Encryption Key field with only method in
// "k=<method>" format to Session.
func (s Session) AddEncryptionMethod(method string) Session {
	return s.appendString(TypeEncryptionKeys, method)
}

func getDefault(v, d string) string {
	if len(v) == 0 {
		return d
	}
	return v
}
