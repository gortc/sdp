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
