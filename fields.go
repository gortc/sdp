package sdp

import (
	"net"
	"strconv"
	"strings"
)

func appendSpace(v []byte) []byte {
	return appendRune(v, ' ')
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

func (s Session) AddConnectionData(netType, addType, data string) Session {
	v := make([]byte, 0, 256)
	v = append(v, netType...)
	v = appendSpace(v)
	v = append(v, addType...)
	v = appendSpace(v)
	v = append(v, data...)
	return s.append(TypeConnectionData, v)
}

func (s Session) AddConnectionDataIP(ip net.IP) Session {
	netType := "IN"
	addrType := "IP4"
	if len(ip) == net.IPv6len {
		addrType = "IP6"
	}
	ipUpper := strings.ToUpper(ip.String())
	return s.AddConnectionData(netType, addrType, ipUpper)
}
