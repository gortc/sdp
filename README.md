[![Build Status](https://travis-ci.com/gortc/sdp.svg?branch=master)](https://travis-ci.com/gortc/sdp)
[![Master status](https://tc.gortc.io/app/rest/builds/buildType:(id:sdp_MasterStatus)/statusIcon.svg)](https://tc.gortc.io/project.html?projectId=sdp&tab=projectOverview&guest=1)
[![Build status](https://ci.appveyor.com/api/projects/status/gcxr3fq9ebadmu9b?svg=true)](https://ci.appveyor.com/project/ernado/sdp)
[![GoDoc](https://godoc.org/github.com/gortc/sdp?status.svg)](https://godoc.org/github.com/gortc/sdp)
[![codecov](https://codecov.io/gh/gortc/sdp/branch/master/graph/badge.svg)](https://codecov.io/gh/gortc/sdp)
[![stability-beta](https://img.shields.io/badge/stability-beta-33bbff.svg)](https://github.com/mkenney/software-guides/blob/master/STABILITY-BADGES.md#beta)
[![Go Report Card](https://goreportcard.com/badge/github.com/gortc/sdp)](https://goreportcard.com/report/github.com/gortc/sdp)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fgortc%2Fsdp.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fgortc%2Fsdp?ref=badge_shield)

# SDP
Package sdp implements SDP: Session Description Protocol [[RFC4566](https://tools.ietf.org/html/rfc4566)].
Complies to [gortc principles](https://gortc.io/#principles) as core package.

### Examples
See [examples](https://github.com/gortc/sdp/tree/master/examples) folder.
Also there is [online SDP example](https://gortc.io/x/sdp/) that gets
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
			Formats:   []string{"0"},
			Protocol: "RTP/AVP",
		},
	}
	video := sdp.Media{
		Description: sdp.MediaDescription{
			Type:     "video",
			Port:     51372,
			Formats:   []string{"99"},
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
			Formats:   []string{"99"},
			Protocol: "RTP/AVP",
		}).
		AddAttribute("rtpmap", "99", "h263-1998/90000").
		AddLine(sdp.TypeEmail, "test@test.com").
		AddRaw('ü', "vαlue").
		AppendTo(b)
	// and so on
	fmt.Println(string(b))
	// Output:
	//	v=0
	//	m=video 51372 RTP/AVP 99
	//	a=rtpmap:99 h263-1998/90000
	//	e=test@test.com
	//	ü=vαlue
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

### Benchmarks
```
goos: linux
goarch: amd64
pkg: github.com/gortc/sdp
PASS
benchmark                                    iter       time/iter   bytes alloc         allocs
---------                                    ----       ---------   -----------         ------
BenchmarkDecoder_Decode-12                 300000   4884.00 ns/op     3166 B/op   93 allocs/op
BenchmarkEncode-12                        1000000   1577.00 ns/op        0 B/op    0 allocs/op
BenchmarkSession_AddConnectionData-12    20000000    114.00 ns/op        0 B/op    0 allocs/op
BenchmarkAppendIP-12                     50000000     37.90 ns/op        0 B/op    0 allocs/op
BenchmarkAppendByte-12                  100000000     11.00 ns/op        0 B/op    0 allocs/op
BenchmarkAppendInt-12                   100000000     11.90 ns/op        0 B/op    0 allocs/op
BenchmarkSession_EX1-12                   3000000    578.00 ns/op       16 B/op    1 allocs/op
BenchmarkAppendRune-12                  200000000      6.70 ns/op        0 B/op    0 allocs/op
BenchmarkDecode-12                      100000000     13.10 ns/op        0 B/op    0 allocs/op
BenchmarkDecodeSession-12                 5000000    234.00 ns/op        0 B/op    0 allocs/op
ok  	github.com/gortc/sdp	16.820s
```

## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fgortc%2Fsdp.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fgortc%2Fsdp?ref=badge_large)