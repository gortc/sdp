package sdp

import (
	"log"
	"os"
	"testing"

	"github.com/pkg/errors"
)

func TestDecoder_Decode(t *testing.T) {
	m := new(Message)
	tData := loadData(t, "sdp_session_ex_full")
	session, err := DecodeSession(tData, nil)
	if err != nil {
		t.Fatal(err)
	}
	decoder := Decoder{
		s: session,
	}
	if err := decoder.Decode(m); err != nil {
		errors.Fprint(os.Stderr, err)
		t.Error(err)
	}
	if m.Version != 0 {
		t.Error("wat", m.Version)
	}
	if !m.Flag("recvonly") {
		t.Error("flag recvonly not found")
	}
	if len(m.Medias) != 2 {
		t.Error("len(medias)", len(m.Medias))
	}
	if m.Medias[1].Attributes.Value("rtpmap") != "99 h263-1998/90000" {
		log.Println(m.Medias[1].Attributes)
		log.Println(m.Medias[0].Attributes)
		t.Error("rtpmap", m.Medias[1].Attributes.Value("rtpmap"))
	}
	if m.Bandwidth != 154798 {
		t.Error("bandwidth bad value", m.Bandwidth)
	}
	if m.BandwidthType != BandwidthConferenceTotal {
		t.Error("bandwidth bad type", m.BandwidthType)
	}
	expectedEncryption := Encryption{"clear", "ab8c4df8b8f4as8v8iuy8re"}
	if m.Encryption != expectedEncryption {
		t.Error("bad encryption", m.Encryption, "!=", expectedEncryption)
	}
	expectedEncryption = Encryption{Method: "prompt"}
	if m.Medias[1].Encryption != expectedEncryption {
		t.Error("bad encryption",
			m.Medias[1].Encryption, "!=", expectedEncryption)
	}
}
