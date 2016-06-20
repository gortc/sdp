package sdp

func (s Session) appendAttributes(attrs Attributes) Session {
	for k, v := range attrs {
		for _, a := range v {
			if len(a) == 0 {
				s = s.AddFlag(k)
			} else {
				s = s.AddAttribute(k, a)
			}
		}
	}
	return s
}

// Append encodes message to Session and returns result.
func (m *Message) Append(s Session) Session {
	// see https://tools.ietf.org/html/rfc4566#section-5
	s = s.AddVersion(m.Version)
	s = s.AddOrigin(m.Origin)
	s = s.AddSessionName(m.Name)
	if len(m.Info) > 0 {
		s = s.AddSessionInfo(m.Info)
	}
	if len(m.URI) > 0 {
		s = s.AddURI(m.URI)
	}
	if len(m.Email) > 0 {
		s = s.AddEmail(m.Email)
	}
	if len(m.Phone) > 0 {
		s = s.AddPhone(m.Phone)
	}
	if !m.Connection.Blank() {
		s = s.AddConnectionData(m.Connection)
	}
	for t, v := range m.Bandwidths {
		s = s.AddBandwidth(t, v)
	}
	// One or more time descriptions ("t=" and "r=" lines)
	for _, t := range m.Timing {
		s = s.AddTiming(t.Start, t.End)
		if len(t.Offsets) > 0 {
			s = s.AddRepeatTimesCompact(t.Repeat, t.Active, t.Offsets...)
		}
	}
	if len(m.TZAdjustments) > 0 {
		s = s.AddTimeZones(m.TZAdjustments...)
	}
	if !m.Encryption.Blank() {
		s = s.AddEncryption(m.Encryption)
	}
	s = s.appendAttributes(m.Attributes)

	for _, mm := range m.Medias {
		s = s.AddMediaDescription(mm.Description)
		if len(mm.Title) > 0 {
			s = s.AddSessionInfo(mm.Title)
		}
		if !mm.Connection.Blank() {
			s = s.AddConnectionData(mm.Connection)
		}
		for t, v := range mm.Bandwidths {
			s = s.AddBandwidth(t, v)
		}
		if !mm.Encryption.Blank() {
			s = s.AddEncryption(mm.Encryption)
		}
		s = s.appendAttributes(mm.Attributes)
	}
	return s
}
