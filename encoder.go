package sdp

// Append encodes message to Session and returns result.
func (m *Message) Append(s Session) Session {
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
	if len(m.TZAdjustments) > 0 {
		s = s.AddTimeZones(m.TZAdjustments...)
	}
	for _, a := range m.Attributes {
		if len(a) == 1 {
			s = s.AddFlag(a[0])
		} else {
			s = s.AddAttribute(a[0], a[1:]...)
		}
	}
	for _, t := range m.Timing {
		s = s.AddTiming(t.Start, t.End)
		if len(t.Offsets) > 0 {
			s = s.AddRepeatTimesCompact(t.Repeat, t.Active, t.Offsets...)
		}
	}
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
			s = s.AddEncryptionKey(mm.Encryption.Method, mm.Encryption.Key)
		}
		for _, a := range mm.Attributes {
			if len(a) == 1 {
				s = s.AddFlag(a[0])
			} else {
				s = s.AddAttribute(a[0], a[1:]...)
			}
		}
	}
	return s
}
