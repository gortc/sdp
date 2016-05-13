package sdp

import "testing"

func BenchmarkAppendRune(b *testing.B) {
	b.ReportAllocs()
	buf := make([]byte, 8)
	r := rune(TypeAttributes)
	for i := 0; i < b.N; i++ {
		buf = appendRune(buf, r)
		buf = buf[:0]
	}
}

func TestEncodeDecode(t *testing.T) {
	v := Line{
		Type:  TypeOrigin,
		Value: []byte("origin?"),
	}
	buf := make([]byte, 0, 128)
	buf = v.AppendTo(buf)
	decoded := Line{}
	if err := decoded.Decode(buf); err != nil {
		t.Error(err)
	}
}

func TestLineDecode(t *testing.T) {
	var tests = []struct {
		s        string
		expected Line
	}{
		{
			"a=value",
			Line{TypeAttributes, []byte("value")},
		},
		{
			"б=значение",                  // unknown attribute char
			Line{'б', []byte("значение")}, // unicode.
			// btw, б is russian "b" and значение is "value" in english.
			// non-english characters are not common for SDP, but we must
			// handle it, because it is UTF.
		},
	}
	for i, tt := range tests {
		actual := Line{}
		if err := actual.Decode([]byte(tt.s)); err != nil {
			t.Errorf("tt[%d]: %v", i, err)
		}
		if !actual.Equal(tt.expected) {
			t.Errorf("tt[%d]: %s != %s", i, actual, tt.expected)
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	b.ReportAllocs()
	v := Line{
		Type:  TypeOrigin,
		Value: []byte("origin?"),
	}
	buf := make([]byte, 0, 128)
	buf = v.AppendTo(buf)
	decoded := Line{
		Value: make([]byte, 0, 128),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := decoded.Decode(buf); err != nil {
			b.Fatal(err)
		}
		decoded.Value = decoded.Value[:0]
		decoded.Type = Type(0)
	}
}

func TestType_String(t *testing.T) {
	v := []Type{
		TypeProtocolVersion,
		TypeOrigin,
		TypeSessionName,
		TypeSessionInformation,
		TypeURI,
		TypeEmail,
		TypePhone,
		TypeConnectionData,
		TypeBandwidth,
		TypeTiming,
		TypeRepeatTimes,
		TypeTimeZones,
		TypeEncryptionKeys,
		TypeAttributes,
		TypeMediaDescriptions,
	}
	for _, tt := range v {
		if len(tt.String()) < 2 {
			t.Errorf("Type.String() %s incorrect", tt)
		}
	}

	// unknown type should be printed "as is"
	tt := Type('б')
	if tt.String() != "б" {
		t.Errorf("Type.String() %s != б", tt)
	}
}

func TestLine_String(t *testing.T) {
	var tests = []struct {
		l        Line
		expected string
	}{
		{
			Line{TypeAttributes, []byte("value")},
			"attributes: value",
		},
		{
			Line{'б', []byte("значение")}, // unicode
			"б: значение",                 // unknown attribute char
		},
	}
	for _, tt := range tests {
		actual := tt.l.String()
		if actual != tt.expected {
			t.Errorf("Line.String() %s != %s", actual, tt.expected)
		}
	}
}
