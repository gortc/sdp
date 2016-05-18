package sdp

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func shouldDecode(tb testing.TB, s Session, name string) {
	buf := make([]byte, 0, 1024)
	tData := loadData(tb, name)
	buf = s.AppendTo(buf)
	if !byteSliceEqual(tData, buf) {
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
		IP:             net.ParseIP("10.47.16.5"),
	})
	s = s.AddOrigin(Origin{
		Username:       "jdoe",
		SessionID:      2890844527,
		SessionVersion: 2890842807,
		IP:             net.ParseIP("FF15::103"),
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
		AddAttribute("recvonly").
		AddAttribute("orient", "landscape")
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
