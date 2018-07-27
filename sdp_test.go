package sdp

import (
	"bufio"
	"os"
	"path/filepath"
	"testing"
)

// testdata examples
const (
	dataSDPExample1 = "sdp_session_ex1"
)

var (
	testNL       = []byte{'\n'}
	testCRNL     = []byte{'\r', '\n'}
	testEndLines = []struct {
		name  string
		bytes []byte
	}{
		{"NL", testNL},
		{"CRNL", testCRNL},
	}
)

func loadData(tb testing.TB, name string, newLineBytes []byte) []byte {
	name = filepath.Join("testdata", name+".txt")
	f, err := os.Open(name)
	if err != nil {
		tb.Fatal(err)
	}
	defer func() {
		if errClose := f.Close(); errClose != nil {
			tb.Fatal(errClose)
		}
	}()
	s := bufio.NewScanner(f)
	var v []byte
	nl := false
	for s.Scan() {
		b := s.Bytes()
		if nl {
			v = append(v, newLineBytes...)
		} else {
			nl = true
		}
		v = append(v, b...)
	}
	if err := s.Err(); err != nil {
		tb.Fatal(err)
	}
	return v
}

func BenchmarkAppendRune(b *testing.B) {
	b.ReportAllocs()
	buf := make([]byte, 8)
	r := rune(TypeAttribute)
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
			Line{TypeAttribute, []byte("value")},
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

func TestDecodeSession2(t *testing.T) {
	for _, testEndLine := range testEndLines {
		t.Run(testEndLine.name, func(t *testing.T) {
			data := loadData(t, dataSDPExample1, testEndLine.bytes)
			s, err := DecodeSession(data, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(s) != 12 {
				t.Fatal("length should be 12")
			}
		})
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
		TypeEncryptionKey,
		TypeAttribute,
		TypeMediaDescription,
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
			Line{TypeAttribute, []byte("value")},
			"attribute: value",
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

func TestScanner(t *testing.T) {
	in := []byte(`
               1
2


   3
`)
	counter := 0
	scanner := newScanner(in)
	for scanner.Scan() {
		counter++
		if counter > 3 {
			t.Fatal("too much lines")
		}
	}
	if counter != 3 {
		t.Fatalf("bad length: %d", counter)
	}
}

func TestDecodeSession(t *testing.T) {
	in := ` a=12
	b=41231ar


	б=значение  `
	s, err := DecodeSession([]byte(in), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 3 {
		t.Fatalf("len(s) != 3, but %d", len(s))
	}
}

func BenchmarkDecodeSession(b *testing.B) {
	in := []byte(` a=12
	b=41231ar


	б=значение  `)
	session := make(Session, 0, 15)
	var err error
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		session, err = DecodeSession(in, session)
		if err != nil {
			b.Fatal(err)
		}
		session = session[:0]
	}
}

func TestDecodeSession_Errors(t *testing.T) {
	shouldFail := []struct {
		Name string
		Data string
	}{
		{"No delimitor", "v"},
		{"No value", "v="},
	}
	var (
		s   Session
		err error
	)
	for _, test := range shouldFail {
		t.Run(test.Name, func(t *testing.T) {
			s, err = DecodeSession([]byte(test.Data), s)
			if err == nil {
				t.Errorf("should fail")
			}
		})
	}
}

func TestSession_Equal(t *testing.T) {
	for _, tc := range []struct {
		a, b  Session
		name  string
		value bool
	}{
		{
			name:  "blank",
			value: true,
		},
		{
			name:  "length",
			a:     Session{}.AddRaw('a', "b"),
			value: false,
		},
		{
			name:  "value",
			a:     Session{}.AddRaw('a', "b"),
			b:     Session{}.AddRaw('a', "a"),
			value: false,
		},
		{
			name:  "type",
			a:     Session{}.AddRaw('a', "b"),
			b:     Session{}.AddRaw('b', "a"),
			value: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if v := tc.a.Equal(tc.b); v != tc.value {
				t.Error("unexpected value")
			}
		})
	}
}

func TestDecodeError_Error(t *testing.T) {
	if newDecodeError("p", "r").Error() != "DecodeError in p: r" {
		t.Error("bad decode error description")
	}
}
