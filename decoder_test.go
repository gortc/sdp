package sdp

import (
	"fmt"
	"log"
	"net"
	"testing"
	"time"
)

func TestDecodeInterval(t *testing.T) {
	var dt = []struct {
		in         string
		out        time.Duration
		shouldFail bool
	}{
		{"0", 0, false},
		{"7d", time.Hour * 24 * 7, false},
		{"25h", time.Hour * 25, false},
		{"45s", time.Second * 45, false},
		{"19m", time.Minute * 19, false},
		{"63050", time.Second * 63050, false},
		{"5", time.Second * 5, false},
		{"s", time.Second * 0, true},
		{"zs", time.Second * 0, true},
	}
	for _, dtt := range dt {
		t.Run(dtt.in, func(t *testing.T) {
			var (
				d time.Duration
				v = []byte(dtt.in)
			)
			err := decodeInterval(v, &d)
			if dtt.shouldFail {
				if err == nil {
					t.Fatal("should fail")
				}
			} else {
				if err != nil {
					t.Errorf("err: %s", err)
				}
				if d != dtt.out {
					t.Errorf("wrong output: %s", dtt.out)
				}
			}
		})
	}
}

func TestDecoder_Decode(t *testing.T) {
	for _, testEndLine := range testEndLines {
		t.Run(testEndLine.name, func(t *testing.T) {
			m := new(Message)
			tData := loadData(t, "sdp_session_ex_full", testEndLine.bytes)
			session, err := DecodeSession(tData, nil)
			if err != nil {
				t.Fatal(err)
			}
			decoder := NewDecoder(session)
			if err := decoder.Decode(m); err != nil {
				t.Fatal(err)
			}
			if m.Version != 0 {
				t.Error("wat", m.Version)
			}
			if !m.Flag("recvonly") {
				t.Error("flag recvonly not found")
			}
			if len(m.Medias) != 2 {
				t.Error("len(medias)", len(m.Medias))
			}
			if m.Medias[1].Attributes.Value("rtpmap") != "99 h263-1998/90000" {
				log.Println(m.Medias[1].Attributes)
				log.Println(m.Medias[0].Attributes)
				t.Error("rtpmap", m.Medias[1].Attributes.Value("rtpmap"))
			}
			if m.Bandwidths[BandwidthConferenceTotal] != 154798 {
				t.Error("bandwidth bad value", m.Bandwidths[BandwidthConferenceTotal])
			}
			expectedEncryption := Encryption{"clear", "ab8c4df8b8f4as8v8iuy8re"}
			if m.Encryption != expectedEncryption {
				t.Error("bad encryption", m.Encryption, "!=", expectedEncryption)
			}
			expectedEncryption = Encryption{Method: "prompt"}
			if m.Medias[1].Encryption != expectedEncryption {
				t.Error("bad encryption",
					m.Medias[1].Encryption, "!=", expectedEncryption)
			}
			if m.Start() != NTPToTime(2873397496) {
				t.Error(m.Start(), "!=", NTPToTime(2873397496))
			}
			if m.End() != NTPToTime(2873404696) {
				t.Error(m.End(), "!=", NTPToTime(2873404696))
			}
			cExpected := ConnectionData{
				IP:          net.ParseIP("224.2.17.12"),
				TTL:         127,
				NetworkType: "IN",
				AddressType: "IP4",
			}
			if !cExpected.Equal(m.Connection) {
				t.Error(cExpected, "!=", m.Connection)
			}
			// jdoe 2890844526 2890842807 IN IP4 10.47.16.5
			oExpected := Origin{
				Address:        "10.47.16.5",
				SessionID:      2890844526,
				SessionVersion: 2890842807,
				NetworkType:    "IN",
				AddressType:    "IP4",
				Username:       "jdoe",
			}
			if !oExpected.Equal(m.Origin) {
				t.Error(oExpected, "!=", m.Origin)
			}
			if len(m.TZAdjustments) < 2 {
				t.Error("tz adjustments count unexpected")
			} else {
				tzExp := TimeZone{
					Start:  NTPToTime(2882844526),
					Offset: -1 * time.Hour,
				}
				if m.TZAdjustments[0] != tzExp {
					t.Error("tz", m.TZAdjustments[0], "!=", tzExp)
				}
				tzExp = TimeZone{
					Start:  NTPToTime(2898848070),
					Offset: 0,
				}
				if m.TZAdjustments[1] != tzExp {
					t.Error("tz", m.TZAdjustments[0], "!=", tzExp)
				}
			}

			if len(m.Medias) != 2 {
				t.Error("media count unexpected")
			} else {
				// audio 49170 RTP/AVP 0
				mExp := MediaDescription{
					Type:     "audio",
					Port:     49170,
					Protocol: "RTP/AVP",
					Formats:  []string{"0"},
				}
				if !m.Medias[0].Description.Equal(mExp) {
					t.Error("m", m.Medias[0].Description, "!=", mExp)
				}
				// video 51372 RTP/AVP 99
				mExp = MediaDescription{
					Type:     "video",
					Port:     51372,
					Protocol: "RTP/AVP",
					Formats:  []string{"99"},
				}
				if !m.Medias[1].Description.Equal(mExp) {
					t.Error("m", m.Medias[1].Description, "!=", mExp)
				}
				if m.Medias[1].PayloadFormat("99") != "h263-1998/90000" {
					t.Error("incorrect payload  format")
				}
				if m.Medias[1].PayloadFormat("0") != "" {
					t.Error("incorrect payload  format")
				}
			}
		})
	}
}

func TestDecoder_WebRTC1(t *testing.T) {
	for _, testEndLine := range testEndLines {
		t.Run(testEndLine.name, func(t *testing.T) {
			m := new(Message)
			tData := loadData(t, "spd_session_ex_webrtc1", testEndLine.bytes)
			session, err := DecodeSession(tData, nil)
			if err != nil {
				t.Fatal(err)
			}
			//for _, l := range session {
			//	fmt.Println(l)
			//}
			decoder := Decoder{
				s: session,
			}
			if err := decoder.Decode(m); err != nil {
				t.Error(err)
			}
			if m.Version != 0 {
				t.Error("wat", m.Version)
			}
		})
	}
}

func BenchmarkDecoder_Decode(b *testing.B) {
	m := new(Message)
	b.ReportAllocs()
	tData := loadData(b, "sdp_session_ex_full", testNL)
	session, err := DecodeSession(tData, nil)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoder := Decoder{
			s: session,
		}
		if err := decoder.Decode(m); err != nil {
			b.Error(err)
		}
		m.Medias = m.Medias[:0]
	}
}

func TestMessage_Empty(t *testing.T) {
	m := &Message{}
	if m.Flag("noFlag") {
		t.Error("flag found")
	}
	if m.Attribute("noAttribute") != blank {
		t.Errorf("attribute found")
	}
}

func TestMessage_Defaults(t *testing.T) {
	m := &Message{}
	m.AddAttribute("k", "v")
	m.AddFlag("readonly")
	if !m.Flag("readonly") {
		t.Error("flag readonly not found")
	}
	if len(m.Attributes.Values("k")) != 1 {
		t.Error("len(attrs[k]) != 1")
	}
	if len(m.Attributes.Values("t")) != 0 {
		t.Error("len(attrs[t]) != 0")
	}
	if m.Attributes.Value("t") != blank {
		t.Error("attrs[t] != blank")
	}
	if !m.Start().IsZero() {
		t.Error("m.Start != zero")
	}
	if !m.End().IsZero() {
		t.Error("m.End != zero")
	}
	if m.Flag("t") {
		t.Error("flag t should not be true")
	}
	if m.Attribute("k") != "v" {
		t.Errorf("attrs[k] %s != v", m.Attribute("k"))
	}
}

func TestMedia_Defaults(t *testing.T) {
	m := &Media{}
	if m.Flag("readonly") {
		t.Error("flag readonly should not be true")
	}
	m.AddAttribute("k", "v")
	m.AddFlag("readonly")
	if !m.Flag("readonly") {
		t.Error("flag readonly not found")
	}
	if len(m.Attributes.Values("k")) != 1 {
		t.Error("len(attrs[k]) != 1")
	}
	if len(m.Attributes.Values("t")) != 0 {
		t.Error("len(attrs[t]) != 0")
	}
	if m.Attributes.Value("t") != blank {
		t.Error("attrs[t] != blank")
	}
	if m.Flag("t") {
		t.Error("flag t should not be true")
	}
	if m.Attribute("k") != "v" {
		t.Errorf("attrs[k] %s != v", m.Attribute("k"))
	}
}

func TestDecoder_Errors(t *testing.T) {
	shouldFail := []string{
		"sdp_session_ex_err1",  // Bandwidth: In time description
		"sdp_session_ex_err2",  // Origin: To many sub-fields
		"sdp_session_ex_err3",  // Invalid text
		"sdp_session_ex_err4",  // Bandwidth: In time description
		"sdp_session_ex_err5",  // ConnectionData: No connectionAddress
		"sdp_session_ex_err6",  // RepeatTimes: Repeat times without times
		"sdp_session_ex_err7",  // RepeatTimes: Double space
		"sdp_session_ex_err8",  // SessionName: Missing
		"sdp_session_ex_err9",  // Origin: Missing
		"sdp_session_ex_err10", // RepeatTimes: Starting with a space
		"sdp_session_ex_err11", // Attribute: decodeKV: Attribute without value
		"sdp_session_ex_err12", // EncryptionKey: decodeKV: Attribute without value
		"sdp_session_ex_err13", // Bandwidth: decodeKV: Attribute without value
		"sdp_session_ex_err14", // ProtocolVersion: Not a number
		"sdp_session_ex_err15", // MediaDescription: < 4 sub-fields
		"sdp_session_ex_err17", // ConnectionData: No addressType
		"sdp_session_ex_err18", // ConnectionData: To many sub-fields
		"sdp_session_ex_err19", // ConnectionData: To many connectionAddress  sub-fields
		"sdp_session_ex_err20", // ConnectionData: Unexpected TTL for IPv6
		"sdp_session_ex_err21", // ConnectionData: Invalid TTL
		"sdp_session_ex_err22", // ConnectionData: Invalid number of addresses
		"sdp_session_ex_err23", // ConnectionData: Invalid number of addresses with ttl
		"sdp_session_ex_err24", // ConnectionData: Invalid number of addresses IPV6
		"sdp_session_ex_err26", // Bandwidth: Invalid BW type
		"sdp_session_ex_err27", // Bandwidth: Invalid value
		"sdp_session_ex_err28", // Timing: To many sub-fields
		"sdp_session_ex_err29", // Timing: Invalid start
		"sdp_session_ex_err30", // Timing: Invalid end
		"sdp_session_ex_err31", // Origin: Invalid subfields
		"sdp_session_ex_err32", // Origin: Invalid sess-id
		"sdp_session_ex_err33", // Origin: Invalid sess-version
		"sdp_session_ex_err34", // RepeatTimes: Invalid number of sub-fields
		"sdp_session_ex_err35", // RepeatTimes: Invalid repeat interval
		"sdp_session_ex_err36", // RepeatTimes: Invalid active duration
		"sdp_session_ex_err37", // RepeatTimes: Invalid offsets from start-time
		"sdp_session_ex_err38", // TimeZones: Invalid double space
		"sdp_session_ex_err39", // TimeZones: Invalid number of sub-fields
		"sdp_session_ex_err40", // TimeZones: Invalid offset
		"sdp_session_ex_err41", // MediaDescription: Invalid double space
		"sdp_session_ex_err42", // MediaDescription: Invalid port
		"sdp_session_ex_err43", // MediaDescription: Invalid number of ports
		"sdp_session_ex_err46", // ConnectionData: Invalid TTL in media secion
		"sdp_session_ex_err47", // ConnectionData: Invalid number of addresses in media secion
		"sdp_session_ex_err48", // ConnectionData: Invalid number of addresses with ttl in media secion
		"sdp_session_ex_err49", // ConnectionData: Invalid number of addresses IPV6 in media secion
		"sdp_session_ex_err50", // RepeatTimes: Unexpected transition
		"sdp_session_ex_err51", // Media: Unexpected transition
		"sdp_session_ex_err52", // ConnectionData: Invalid IPv4
		"sdp_session_ex_err53", // Bandwidth: not K-V pair
	}
	var (
		s   Session
		err error
	)
	for _, name := range shouldFail {
		t.Run(name, func(t *testing.T) {
			b := loadData(t, name, testNL)
			s, err = DecodeSession(b, s)
			if err != nil {
				t.Fatalf("err: %s", err)
			}
			m := new(Message)
			d := NewDecoder(s)
			err = d.Decode(m)
			s = s.reset()
			if err == nil {
				t.Errorf("should fail")
			}
		})
	}
}

func TestDecoder_ExMediaConnection(t *testing.T) {
	m := new(Message)
	tData := loadData(t, "sdp_session_ex_mediac", testNL)
	session, err := DecodeSession(tData, nil)
	if err != nil {
		t.Fatal(err)
	}
	decoder := NewDecoder(session)
	if err := decoder.Decode(m); err != nil {
		t.Fatal(err)
	}
	if len(m.Medias) != 2 {
		t.Error("media count unexpected")
	} else {
		cExpected := ConnectionData{
			IP:          net.ParseIP("0.0.0.0"),
			NetworkType: "IN",
			AddressType: "IP4",
		}
		if got := m.Medias[0].Connection; !cExpected.Equal(got) {
			t.Errorf("%s (got) != %s (expected)", got, cExpected)
		}
	}
}

func TestDecoder_NoMediaFmt(t *testing.T) {
	m := new(Message)
	tData := loadData(t, "sdp_session_no_media_fmt", testNL)
	session, err := DecodeSession(tData, nil)
	if err != nil {
		t.Fatal(err)
	}
	decoder := NewDecoder(session)
	if err := decoder.Decode(m); err != nil {
		t.Fatal(err)
	}
}

func TestSectionOverflows(t *testing.T) {
	mustOverflow := func(t *testing.T) {
		if err := recover(); err != "BUG: section overflow" {
			t.Error("should panic")
		}
	}
	t.Run("String", func(t *testing.T) {
		defer mustOverflow(t)
		fmt.Print(section(123).String())
	})
	t.Run("Ordering", func(t *testing.T) {
		defer mustOverflow(t)
		fmt.Print(getOrdering(section(123)))
	})
}

func TestDecoderUnexpectedField(t *testing.T) {
	mustBug := func(t *testing.T) {
		if err := recover(); err != "BUG: unexpected filed type in decodeField" {
			t.Error("should panic")
		}
	}
	d := NewDecoder(Session{})
	t.Run("ShouldPanic", func(t *testing.T) {
		defer mustBug(t)
		d.t = Type('1')
		d.decodeField(nil)
	})
}

func TestDecoderNettypeEmpty(t *testing.T) {
	d := NewDecoder(Session{})
	m := &Message{}
	if err := d.decodeConnectionData(m); err == nil {
		t.Error("should error")
	}
}

func TestDecodeLastTiming(t *testing.T) {
	m := new(Message)
	tData := loadData(t, "sdp_session_ex_lasttiming", testNL)
	session, err := DecodeSession(tData, nil)
	if err != nil {
		t.Fatal(err)
	}
	decoder := NewDecoder(session)
	if err := decoder.Decode(m); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeUnknownType(t *testing.T) {
	m := new(Message)
	tData := loadData(t, "sdp_session_ex_media_unknown_type", testNL)
	session, err := DecodeSession(tData, nil)
	if err != nil {
		t.Fatal(err)
	}
	decoder := NewDecoder(session)
	if err := decoder.Decode(m); err != nil {
		t.Fatal(err)
	}
}
