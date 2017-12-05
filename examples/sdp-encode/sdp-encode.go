package main

import (
	"fmt"
	"net"
	"time"

	"github.com/gortc/sdp"
)

func main() {
	var (
		s sdp.Session
		b []byte
	)
	// defining medias
	audio := sdp.Media{
		Description: sdp.MediaDescription{
			Type:     "audio",
			Port:     49170,
			Format:   "0",
			Protocol: "RTP/AVP",
		},
	}
	video := sdp.Media{
		Description: sdp.MediaDescription{
			Type:     "video",
			Port:     51372,
			Format:   "99",
			Protocol: "RTP/AVP",
		},
		Bandwidths: sdp.Bandwidths{
			sdp.BandwidthApplicationSpecific: 66781,
		},
		Encryption: sdp.Encryption{
			Method: "prompt",
		},
	}
	video.AddAttribute("rtpmap", "99", "h263-1998/90000")

	// defining message
	m := &sdp.Message{
		Origin: sdp.Origin{
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
		Connection: sdp.ConnectionData{
			IP:  net.ParseIP("224.2.17.12"),
			TTL: 127,
		},
		Bandwidths: sdp.Bandwidths{
			sdp.BandwidthConferenceTotal: 154798,
		},
		Timing: []sdp.Timing{
			{
				Start:  sdp.NTPToTime(2873397496),
				End:    sdp.NTPToTime(2873404696),
				Repeat: 7 * time.Hour * 24,
				Active: 3600 * time.Second,
				Offsets: []time.Duration{
					0,
					25 * time.Hour,
				},
			},
		},
		Encryption: sdp.Encryption{
			Method: "clear",
			Key:    "ab8c4df8b8f4as8v8iuy8re",
		},
		Medias: []sdp.Media{audio, video},
	}
	m.AddFlag("recvonly")

	// appending message to session
	s = m.Append(s)

	// appending session to byte buffer
	b = s.AppendTo(b)
	fmt.Println(string(b))
}
