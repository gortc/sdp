package sdp

import "strconv"

func (s Session) AddVersion(v int) Session {
	return s.append(TypeProtocolVersion, []byte(strconv.Itoa(v)))
}
