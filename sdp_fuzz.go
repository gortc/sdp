// +build gofuzz

package sdp

import "fmt"

// Instructions
// 0. Create a folder called 'corpus' with the contents of the 'testdata' folder
// 1. Get
//    go get -u github.com/dvyukov/go-fuzz/...
// 2. Build
//    go-fuzz-build github.com/gortc/sdp
// 3. Run
//    go-fuzz --bin=sdp-fuzz.zip --workdir=fuzz

func Fuzz(data []byte) int {
	s, err := DecodeSession(data, nil)
	if err != nil {
		return 0
	}

	m := new(Message)
	decoder := NewDecoder(s)
	if err := decoder.Decode(m); err != nil {
		return 0
	}

	s2 := make(Session, 0, 100)
	s2 = m.Append(s2)
	buf := make([]byte, 0, 1024)
	buf = s2.AppendTo(buf)

	if len(buf) < 2 {
		return 0
	}

	s3, err := DecodeSession(buf, nil)
	if err != nil {
		return 0
	}

	if !s3.Equal(s2) {
		dbg := make([]byte, 0, 1024)
		dbg = s3.AppendTo(dbg)
		msg := fmt.Sprintf("Not equal: %q != %q", dbg, buf)
		panic(msg)
	}

	return 1
}
