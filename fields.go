package sdp

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func appendSpace(v []byte) []byte {
	return appendRune(v, ' ')
}

func appendInt(v []byte, i int) []byte {
	return append(v, strconv.Itoa(i)...)
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
	return s.AddConnectionDataRaw(
		data.getNetworkType(),
		data.getAddressType(),
		data.ConnectionAddress(),
	)
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
	var address = strings.ToUpper(c.IP.String())
	if c.TTL > 0 {
		address += fmt.Sprintf("/%d", c.TTL)
	}
	if c.Addresses > 0 {
		address += fmt.Sprintf("/%d", c.Addresses)
	}
	return address
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

func getDefault(v, d string) string {
	if len(v) == 0 {
		return d
	}
	return v
}
