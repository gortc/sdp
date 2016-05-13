package sdp

import (
	"testing"
)

func shouldDecode(tb testing.TB, s Session, name string) {
	buf := make([]byte, 0, 1024)
	tData := loadData(tb, name)
	buf = s.AppendTo(buf)
	if !byteSliceEqual(tData, buf) {
		tb.Errorf("%s != %s", string(tData), string(buf))
	}
}

func TestSession_AddVersion(t *testing.T) {
	shouldDecode(t, new(Session).AddVersion(1337), "spd_session_ex2")
}
