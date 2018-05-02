package sdp

import (
	"log"
	"net"
	"testing"
	"time"
)

func TestDecodeInterval(t *testing.T) {
	var dt = []struct {
		in  string
		out time.Duration
	}{
		{"0", 0},
		{"7d", time.Hour * 24 * 7},
		{"25h", time.Hour * 25},
		{"63050", time.Second * 63050},
		{"5", time.Second * 5},
	}
	for i, dtt := range dt {
		var (
			d time.Duration
			v = []byte(dtt.in)
		)
		if err := decodeInterval(v, &d); err != nil {
			t.Errorf("dtt[%d]: dec(%s) err: %s", i, dtt.in, err)
			continue
		}
		if d != dtt.out {
			t.Errorf("dtt[%d]: dec(%s) = %s != %s", i, dtt.in, d, dtt.out)
		}
	}
}

func TestDecoder_Decode(t *testing.T) {
	m := new(Message)
	tData := loadData(t, "sdp_session_ex_full")
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
			Format:   "0",
		}
		if m.Medias[0].Description != mExp {
			t.Error("m", m.Medias[0].Description, "!=", mExp)
		}
		// video 51372 RTP/AVP 99
		mExp = MediaDescription{
			Type:     "video",
			Port:     51372,
			Protocol: "RTP/AVP",
			Format:   "99",
		}
		if m.Medias[1].Description != mExp {
			t.Error("m", m.Medias[1].Description, "!=", mExp)
		}
	}
}

func TestDecoder_WebRTC1(t *testing.T) {
	m := new(Message)
	tData := loadData(t, "spd_session_ex_webrtc1")
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
}

func BenchmarkDecoder_Decode(b *testing.B) {
	m := new(Message)
	b.ReportAllocs()
	tData := loadData(b, "sdp_session_ex_full")
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
		"sdp_session_ex_err1",
		"sdp_session_ex_err2",
		"sdp_session_ex_err3",
		"sdp_session_ex_err4",
		"sdp_session_ex_err5",
	}
	var (
		s   Session
		err error
	)
	for i, name := range shouldFail {
		b := loadData(t, name)
		s, err = DecodeSession(b, s)
		if err != nil {
			t.Fatalf("session %s(%d) err: %s", name, i, err)
		}
		m := new(Message)
		d := NewDecoder(s)
		err = d.Decode(m)
		s = s.reset()
		if err == nil {
			t.Errorf("%s(%d) should fail", name, i)
		}
	}
}

func TestDecoder_ExMediaConnection(t *testing.T) {
	m := new(Message)
	tData := loadData(t, "sdp_session_ex_mediac")
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
