package sdp

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestMessage_Append(t *testing.T) {
	audio := Media{
		Description: MediaDescription{
			Type:     "audio",
			Port:     49170,
			Format:   "0",
			Protocol: "RTP/AVP",
		},
	}
	video := Media{
		Description: MediaDescription{
			Type:     "video",
			Port:     51372,
			Format:   "99",
			Protocol: "RTP/AVP",
		},
		Bandwidths: Bandwidths{
			BandwidthApplicationSpecific: 66781,
		},
		Encryption: Encryption{
			Method: "prompt",
		},
	}
	video.AddAttribute("rtpmap", "99", "h263-1998/90000")

	m := &Message{
		Origin: Origin{
			Username:       "jdoe",
			SessionID:      2890844526,
			SessionVersion: 2890842807,
			Address:        "10.47.16.5",
		},
		Name:  "SDP Seminar",
		Info:  "A Seminar on the session description protocol",
		URI:   "http://www.example.com/seminars/sdp.pdf",
		Email: "j.doe@example.com (Jane Doe)",
		Phone: "12345",
		Connection: ConnectionData{
			IP:  net.ParseIP("224.2.17.12"),
			TTL: 127,
		},
		Bandwidths: Bandwidths{
			BandwidthConferenceTotal: 154798,
		},
		Timing: []Timing{
			{
				Start:  NTPToTime(2873397496),
				End:    NTPToTime(2873404696),
				Repeat: 7 * time.Hour * 24,
				Active: 3600 * time.Second,
				Offsets: []time.Duration{
					0,
					25 * time.Hour,
				},
			},
		},
		Encryption: Encryption{
			Method: "clear",
			Key:    "ab8c4df8b8f4as8v8iuy8re",
		},
		Medias: []Media{audio, video},
	}
	m.AddFlag("recvonly")
	result := `v=0
o=jdoe 2890844526 2890842807 IN IP4 10.47.16.5
s=SDP Seminar
i=A Seminar on the session description protocol
u=http://www.example.com/seminars/sdp.pdf
e=j.doe@example.com (Jane Doe)
p=12345
c=IN IP4 224.2.17.12/127
b=CT:154798
t=2873397496 2873404696
r=7d 1h 0 25h
k=clear:ab8c4df8b8f4as8v8iuy8re
a=recvonly
m=audio 49170 RTP/AVP 0
m=video 51372 RTP/AVP 99
b=AS:66781
k=prompt
a=rtpmap:99 h263-1998/90000`
	s := make(Session, 0, 100)
	s = m.Append(s)
	buf := make([]byte, 0, 1024)
	buf = s.AppendTo(buf)
	if result != string(buf) {
		for k, v := range m.Attributes {
			fmt.Println(k, v)
		}
		for i, v := range s {
			fmt.Println(i, v)
		}
		t.Error(string(buf))
	}
}

func BenchmarkEncode(b *testing.B) {
	b.ReportAllocs()
	audio := Media{
		Description: MediaDescription{
			Type:     "audio",
			Port:     49170,
			Format:   "0",
			Protocol: "RTP/AVP",
		},
	}
	video := Media{
		Description: MediaDescription{
			Type:     "video",
			Port:     51372,
			Format:   "99",
			Protocol: "RTP/AVP",
		},
		Bandwidths: Bandwidths{
			BandwidthApplicationSpecific: 66781,
		},
		Encryption: Encryption{
			Method: "prompt",
		},
	}
	video.AddAttribute("rtpmap", "99", "h263-1998/90000")

	m := &Message{
		Origin: Origin{
			Username:       "jdoe",
			SessionID:      2890844526,
			SessionVersion: 2890842807,
			Address:        "10.47.16.5",
		},
		Name:  "SDP Seminar",
		Info:  "A Seminar on the session description protocol",
		URI:   "http://www.example.com/seminars/sdp.pdf",
		Email: "j.doe@example.com (Jane Doe)",
		Phone: "12345",
		Connection: ConnectionData{
			IP:  net.ParseIP("224.2.17.12"),
			TTL: 127,
		},
		Bandwidths: Bandwidths{
			BandwidthConferenceTotal: 154798,
		},
		Timing: []Timing{
			{
				Start:  NTPToTime(2873397496),
				End:    NTPToTime(2873404696),
				Repeat: 7 * time.Hour * 24,
				Active: 3600 * time.Second,
				Offsets: []time.Duration{
					0,
					25 * time.Hour,
				},
			},
		},
		Encryption: Encryption{
			Method: "clear",
			Key:    "ab8c4df8b8f4as8v8iuy8re",
		},
		Medias: []Media{audio, video},
	}
	m.AddFlag("recvonly")
	s := make(Session, 0, 100)
	buf := make([]byte, 0, 1024)
	for i := 0; i < b.N; i++ {
		s = m.Append(s)
		buf = s.AppendTo(buf)
		s = s.reset()
		buf = buf[:0]
	}
}
