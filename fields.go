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

func appendJoin(b []byte, v [][]byte) []byte {
	last := len(v) - 1
	for i, vv := range v {
		b = append(b, vv...)
		if i != last {
			b = appendSpace(b)
		}
	}
	return b
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

func (s Session) AddVersion(v int) Session {
	return s.append(TypeProtocolVersion, []byte(strconv.Itoa(v)))
}

func (s Session) AddPhone(phone string) Session {
	return s.appendString(TypePhone, phone)
}

func (s Session) AddEmail(email string) Session {
	return s.appendString(TypeEmail, email)
}

func (s Session) AddConnectionData(data ConnectionData) Session {
	return s.AddConnectionDataRaw(
		data.getNetworkType(),
		data.getAddressType(),
		data.ConnectionAddress(),
	)
}

func (s Session) AddConnectionDataRaw(netType, addType, data string) Session {
	v := make([]byte, 0, 256)
	v = appendJoinStrings(v, netType, addType, data)
	return s.append(TypeConnectionData, v)
}

func (s Session) AddConnectionDataIP(ip net.IP) Session {
	return s.AddConnectionData(ConnectionData{
		IP: ip,
	})
}

func (s Session) AddSessionName(name string) Session {
	return s.appendString(TypeSessionName, name)
}

func (s Session) AddSessionInfo(info string) Session {
	return s.appendString(TypeSessionInformation, info)
}

func (s Session) AddURI(uri string) Session {
	return s.appendString(TypeURI, uri)
}

type ConnectionData struct {
	NetworkType string
	AddressType string
	IP          net.IP
	TTL         byte
	Addresses   byte
}

func (c ConnectionData) getNetworkType() string {
	return getDefault(c.NetworkType, "IN")
}

func (c ConnectionData) getAddressType() string {
	if len(c.AddressType) != 0 {
		return c.AddressType
	}
	switch c.IP.To4() {
	case nil:
		return "IP6"
	default:
		return "IP4"
	}
}

func (c ConnectionData) ConnectionAddress() string {
	//  <base multicast address>[/<ttl>]/<number of addresses>
	var address = strings.ToUpper(c.IP.String())
	if c.TTL > 0 {
		address += fmt.Sprintf("/%d", c.TTL)
	}
	if c.Addresses > 0 {
		address += fmt.Sprintf("/%d", c.Addresses)
	}
	return address
}

func getDefault(v, d string) string {
	if len(v) == 0 {
		return d
	}
	return v
}
