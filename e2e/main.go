package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mkenney/go-chrome/tot"
	"github.com/mkenney/go-chrome/tot/runtime"
)

var (
	bin      = flag.String("b", "/usr/bin/google-chrome", "path to binary")
	headless = flag.Bool("headless", true, "headless mode")
	httpAddr = flag.String("addr", "localhost:5568", "http endpoint to listen")
)

func main() {
	flag.Parse()
	gotRequest := make(chan struct{}, 1)
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		log.Println("http:", request.Method, request.RemoteAddr)
		writer.WriteHeader(http.StatusOK)
		fmt.Fprintln(writer, `<h1>Hello world</h1>`)
		gotRequest <- struct{}{}
	})
	go func() {
		if err := http.ListenAndServe(*httpAddr, nil); err != nil {
			log.Fatalln("failed to listen:", err)
		}
	}()
	f := chrome.Flags{}
	if *headless {
		f.Set("headless", "true")
	}
	f.Set("disable-gpu", "true")
	f.Set("disable-dev-shm-usage", "true")
	c := chrome.New(f, *bin, "", "", "")
	if err := c.Launch(); err != nil {
		panic(err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			panic(err)
		}
		log.Println("browser closed")
	}()
	log.Println("navigating to", *httpAddr)
	t, err := c.NewTab(*httpAddr)
	if err != nil {
		log.Fatalln("failed to open tab:", err)
	}
	r := t.Runtime()
	select {
	case <-r.Enable():
		log.Println("runtime enabled")
	case <-time.After(time.Second):
		log.Fatalln("runtime timed out")
	}
	r.OnConsoleAPICalled(func(event *runtime.ConsoleAPICalledEvent) {
		log.Println(event.Type, "in console")
	})
	select {
	case <-gotRequest:
		log.Println("got request")
	case <-time.After(time.Second * 5):
		log.Fatalln("timed out")
	}
}
