package sdp

import (
	"log"
	"os"
	"testing"

	"github.com/pkg/errors"
)

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
}
