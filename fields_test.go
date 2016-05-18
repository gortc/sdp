package sdp

import (
	"fmt"
	"net"
	"testing"
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
			IP:        net.ParseIP("214.6.36.42"),
			TTL:       95,
			Addresses: 4,
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
