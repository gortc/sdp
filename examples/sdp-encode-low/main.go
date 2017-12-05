package main

import (
	"fmt"

	"github.com/gortc/sdp"
)

func main() {
	var (
		s sdp.Session
		b []byte
	)
	b = s.AddVersion(0).
		AddMediaDescription(sdp.MediaDescription{
			Type:     "video",
			Port:     51372,
			Format:   "99",
			Protocol: "RTP/AVP",
		}).
		AddAttribute("rtpmap", "99", "h263-1998/90000").
		AddLine(sdp.TypeEmail, "test@test.com").
		AddRaw('ф', "ОПАСНО").
		AppendTo(b)
	// and so on
	fmt.Println(string(b))
	// Output:
	//	v=0
	//	m=video 51372 RTP/AVP 99
	//	a=rtpmap:99 h263-1998/90000
	//	e=test@test.com
	//  ф=ОПАСНО
}
