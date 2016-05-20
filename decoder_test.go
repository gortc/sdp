package sdp

import "testing"

func TestDecoder_Decode(t *testing.T) {
	m := new(Message)
	tData := loadData(t, "sdp_session_ex1")
	session, err := DecodeSession(tData, nil)
	if err != nil {
		t.Fatal(err)
	}
	decoder := Decoder{
		s: session,
	}
	if err := decoder.Decode(m); err != nil {
		t.Error(err)
	}
	if m.Version != 0 {
		t.Error("wat", m.Version)
	}
}
