package sdp

import (
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"
)

func shouldDecode(tb testing.TB, s Session, name string) {
	buf := make([]byte, 0, 1024)
	tData := loadData(tb, name)
	buf = s.AppendTo(buf)
	if !bytes.Equal(tData, buf) {
		fmt.Println(tData)
		fmt.Println(buf)
		fmt.Println(string(tData))
		fmt.Println(string(buf))
		tb.Errorf("not equal")
	}
	newSesssion, err := DecodeSession(buf, nil)
	if err != nil {
		tb.Errorf("decode error: %v", err)
	}
	if !newSesssion.Equal(s) {
		tb.Error("sessions does not equal")
	}
}

func shouldDecodeExpS(tb testing.TB, s Session, name string) {
	shouldDecode(tb, s, "spd_session_ex_"+name)
}

func TestSession_AddVersion(t *testing.T) {
	shouldDecode(t, new(Session).AddVersion(1337), "spd_session_ex2")
}

func TestSession_AddPhoneEmail(t *testing.T) {
	s := new(Session).AddPhone("+1 617 555-6011")
	s = s.AddEmail("j.doe@example.com (Jane Doe)")
	shouldDecode(t, s, "spd_session_ex3")
}

func TestSession_AddConnectionDataIP(t *testing.T) {
	s := new(Session).
		AddConnectionDataIP(net.ParseIP("ff15::103")).
		AddConnectionData(ConnectionData{
			IP:  net.ParseIP("224.2.36.42"),
			TTL: 127}).
		AddConnectionData(ConnectionData{
			IP:          net.ParseIP("214.6.36.42"),
			AddressType: "IP4",
			NetworkType: "IN",
			TTL:         95,
			Addresses:   4,
		})
	shouldDecodeExpS(t, s, "ip")
}

func TestSession_AddOrigin(t *testing.T) {
	s := new(Session).AddOrigin(Origin{
		Username:       "jdoe",
		SessionID:      2890844526,
		SessionVersion: 2890842807,
		Address:        "10.47.16.5",
	})
	s = s.AddOrigin(Origin{
		Username:       "jdoe",
		SessionID:      2890844527,
		SessionVersion: 2890842807,
		Address:        "FF15::103",
	})
	shouldDecodeExpS(t, s, "origin")
}

func TestSession_AddTiming(t *testing.T) {
	s := new(Session).
		AddTiming(time.Time{}, time.Time{}).
		AddTiming(time.Time{}, time.Unix(833473619, 0))
	shouldDecodeExpS(t, s, "timing")
}

func TestSession_AddAttribute(t *testing.T) {
	s := new(Session).
		AddFlag("recvonly").
		AddAttribute("anotherflag").
		AddAttribute("orient", "landscape").
		AddAttribute("rtpmap", "96", "L8/8000")
	shouldDecodeExpS(t, s, "attributes")
}

func TestSession_AddBandwidth(t *testing.T) {
	s := new(Session).
		AddBandwidth(BandwidthConferenceTotal, 154798).
		AddBandwidth(BandwidthApplicationSpecific, 66781)
	shouldDecodeExpS(t, s, "bandwidth")
}

func TestSession_AddSessionName(t *testing.T) {
	s := new(Session).AddSessionName("CyConf")
	shouldDecodeExpS(t, s, "name")
}

func TestSession_AddSessionInfo(t *testing.T) {
	s := new(Session).AddSessionInfo("Info goes here")
	shouldDecodeExpS(t, s, "info")
}

func TestSession_AddURI(t *testing.T) {
	s := new(Session).AddURI("http://cydev.ru")
	shouldDecodeExpS(t, s, "uri")
}

func TestSession_AddRepeatTimes(t *testing.T) {
	s := new(Session).
		AddRepeatTimes(
			time.Second*604800,
			time.Second*3600,
			0,
			time.Second*90000,
		).
		AddRepeatTimesCompact(
			time.Second*604800,
			time.Second*3600,
			0,
			time.Second*90000,
		).
		AddRepeatTimesCompact(
			time.Second*604810,
			time.Second*3600,
			0,
			time.Second*90000,
		)
	shouldDecodeExpS(t, s, "repeat")
}

func TestSession_AddMediaDescription(t *testing.T) {
	s := new(Session).AddMediaDescription(MediaDescription{
		Type:        "video",
		Port:        49170,
		PortsNumber: 2,
		Protocol:    "RTP/AVP",
		Format:      "31",
	}).AddMediaDescription(MediaDescription{
		Type:     "audio",
		Port:     49170,
		Protocol: "RTP/AVP",
		Format:   "555",
	})
	shouldDecodeExpS(t, s, "media")
}

func TestSession_AddEncryptionKey(t *testing.T) {
	s := new(Session).AddEncryptionKey("clear", "ab8c4df8b8f4as8v8iuy8re").
		AddEncryptionMethod("prompt")
	shouldDecodeExpS(t, s, "keys")
}

func TestSession_AddTimeZones(t *testing.T) {
	s := new(Session).AddTimeZones(
		TimeZone{NTPToTime(2882844526), -1 * time.Hour},
		TimeZone{Start: NTPToTime(2898848070)},
	).AddTimeZones(
		TimeZone{NTPToTime(2898848070), time.Minute * 90},
		TimeZone{Start: NTPToTime(2898848070)},
	)
	shouldDecodeExpS(t, s, "zones")
}

func TestSession_EX1(t *testing.T) {
	/*
		v=0
		o=jdoe 2890844526 2890842807 IN IP4 10.47.16.5
		s=SDP Seminar
		i=A Seminar on the session description protocol
		u=http://www.example.com/seminars/sdp.pdf
		e=j.doe@example.com (Jane Doe)
		c=IN IP4 224.2.17.12/127
		t=2873397496 2873404696
		a=recvonly
		m=audio 49170 RTP/AVP 0
		m=video 51372 RTP/AVP 99
		a=rtpmap:99 h263-1998/90000
	*/

	s := new(Session).
		AddVersion(0).
		AddOrigin(Origin{
			Username:       "jdoe",
			SessionID:      2890844526,
			SessionVersion: 2890842807,
			Address:        "10.47.16.5",
		}).
		AddSessionName("SDP Seminar").
		AddSessionInfo("A Seminar on the session description protocol").
		AddURI("http://www.example.com/seminars/sdp.pdf").
		AddEmail("j.doe@example.com (Jane Doe)").
		AddConnectionData(ConnectionData{
			IP:  net.ParseIP("224.2.17.12"),
			TTL: 127,
		}).
		AddTimingNTP(2873397496, 2873404696).
		AddFlag("recvonly").
		AddMediaDescription(MediaDescription{
			Type:     "audio",
			Port:     49170,
			Protocol: "RTP/AVP",
			Format:   "0",
		}).
		AddMediaDescription(MediaDescription{
			Type:     "video",
			Port:     51372,
			Protocol: "RTP/AVP",
			Format:   "99",
		}).
		AddAttribute("rtpmap", "99", "h263-1998/90000")
	shouldDecode(t, s, "sdp_session_ex1")
}

func BenchmarkSession_AddConnectionData(b *testing.B) {
	s := make(Session, 0, 5)
	b.ReportAllocs()
	var (
		connIP = net.ParseIP("224.2.17.12")
	)
	for i := 0; i < b.N; i++ {
		s = s.AddConnectionData(ConnectionData{
			IP:  connIP,
			TTL: 127,
		})
		s = s.reset()
	}
}

func BenchmarkAppendIP(b *testing.B) {
	buf := make([]byte, 0, 256)
	connIP := net.ParseIP("224.2.17.12")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf = appendIP(buf, connIP)
		buf = buf[:0]
	}
}

func BenchmarkAppendByte(b *testing.B) {
	buf := make([]byte, 0, 64)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf = appendByte(buf, 128)
		buf = buf[:0]
	}
}

func BenchmarkAppendInt(b *testing.B) {
	buf := make([]byte, 0, 64)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf = appendInt(buf, 1024)
		buf = buf[:0]
	}
}

func BenchmarkSession_EX1(b *testing.B) {
	s := make(Session, 0, 30)
	b.ReportAllocs()
	var (
		sessIP = net.ParseIP("10.47.16.5")
		connIP = net.ParseIP("224.2.17.12")
	)
	for i := 0; i < b.N; i++ {
		s = s.AddVersion(0)
		s = s.AddOrigin(Origin{
			Username:       "jdoe",
			SessionID:      2890844526,
			SessionVersion: 2890842807,
			Address:        sessIP.String(),
		})
		s = s.AddSessionName("SDP Seminar")
		s = s.AddSessionInfo("A Seminar on the session description protocol")
		s = s.AddURI("http://www.example.com/seminars/sdp.pdf")
		s = s.AddEmail("j.doe@example.com (Jane Doe)")
		s = s.AddConnectionData(ConnectionData{
			IP:  connIP,
			TTL: 127,
		})
		s = s.AddTimingNTP(2873397496, 2873404696)
		s = s.AddFlag("recvonly")
		s = s.AddMediaDescription(MediaDescription{
			Type:     "audio",
			Port:     49170,
			Protocol: "RTP/AVP",
			Format:   "0",
		})
		s = s.AddMediaDescription(MediaDescription{
			Type:     "video",
			Port:     51372,
			Protocol: "RTP/AVP",
			Format:   "99",
		})
		s = s.AddAttribute("rtpmap", "99", "h263-1998/90000")
		s = s.reset()
	}

}

func TestNTP(t *testing.T) {
	var ntpTable = []struct {
		in  uint64
		out time.Time
	}{
		{3549086042, time.Unix(1340097242, 0)},
		{0, time.Time{}},
	}
	for _, tt := range ntpTable {
		outReal := NTPToTime(tt.in)
		if tt.out != outReal {
			t.Errorf("%v != %v", tt.out, outReal)
		}
		outNTP := TimeToNTP(outReal)
		if outNTP != tt.in {
			t.Errorf("%d != %d", outNTP, tt.in)
		}
	}
}
