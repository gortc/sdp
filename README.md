[![Build Status](https://travis-ci.org/gortc/sdp.svg?branch=master)](https://travis-ci.org/gortc/sdp)
[![Build status](https://ci.appveyor.com/api/projects/status/gcxr3fq9ebadmu9b?svg=true)](https://ci.appveyor.com/project/ernado/sdp)
[![Coverage Status](https://coveralls.io/repos/github/gortc/sdp/badge.svg?branch=master)](https://coveralls.io/github/gortc/sdp?branch=master)
[![GoDoc](https://godoc.org/github.com/gortc/sdp?status.svg)](https://godoc.org/github.com/gortc/sdp)

# SDP go implementation
[RFC 4566](https://tools.ietf.org/html/rfc4566)
SDP: Session Description Protocol in golang.

In alpha stage.

### Examples
See [examples](https://github.com/gortc/sdp/tree/master/examples) folder.
Also there is [online SDP example](https://cydev.ru/sdp/) (temporary unavailable) that gets
`RTCPeerConnection.localDescription.sdp` using WebRTC, 
sends it to server, decodes as `sdp.Session` and renders it on web page.

SDP example:
```sdp
v=0
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
a=rtpmap:99 h263-1998/90000
```
Encode:
```go
package main

import (
	"net"
	"time"
	"fmt"

	"github.com/gortc/sdp"
)

func main()  {
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
			Key: "ab8c4df8b8f4as8v8iuy8re",
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
```
Decode:
```go
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/gortc/sdp"
)

func main() {
	name := "example.sdp"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	var (
		s   sdp.Session
		b   []byte
		err error
		f   io.ReadCloser
	)
	fmt.Println("sdp file:", name)
	if f, err = os.Open(name); err != nil {
		log.Fatal("err:", err)
	}
	defer f.Close()
	if b, err = ioutil.ReadAll(f); err != nil {
		log.Fatal("err:", err)
	}
	if s, err = sdp.DecodeSession(b, s); err != nil {
		log.Fatal("err:", err)
	}
	for k, v := range s {
		fmt.Println(k, v)
	}
	d := sdp.NewDecoder(s)
	m := new(sdp.Message)
	if err = d.Decode(m); err != nil {
		log.Fatal("err:", err)
	}
	fmt.Println("Decoded session", m.Name)
	fmt.Println("Info:", m.Info)
	fmt.Println("Origin:", m.Origin)
}
```
Also, low-level Session struct can be used directly to compose SDP message:
```go
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
```

### Supported params
- [x] v (protocol version)
- [x] o (originator and session identifier)
- [x] s (session name)
- [x] i (session information)
- [x] u (URI of description)
- [x] e (email address)
- [x] p (phone number)
- [x] c (connection information)
- [x] b (zero or more bandwidth information lines)
- [x] t (time)
- [x] r (repeat)
- [x] z (time zone adjustments)
- [x] k (encryption key)
- [x] a (zero or more session attribute lines)
- [x] m (media name and transport address)

### TODO:
- [x] Encoding
- [x] Parsing
- [x] High level encoding
- [x] High level decoding
- [x] Examples
- [x] CI
- [x] More examples and docs
- [x] Online example
- [ ] io.Reader and io.Writer interop
- [ ] Include to high-level CI

### Possible optimizations
There are comments `// ALLOCATIONS: suboptimal.` and `// CPU: suboptimal. `
that indicate suboptimal implementation that can be optimized. There are often
a benchmarks for this pieces.
