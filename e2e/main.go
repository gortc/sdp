package main

import (
	"flag"
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
	gotPostRequest := make(chan struct{}, 1)
	fs := http.FileServer(http.Dir("static"))
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		log.Println("http:", request.Method, request.URL.Path, request.RemoteAddr)
		if request.Method == http.MethodPost {
			gotPostRequest <- struct{}{}
			return
		}
		fs.ServeHTTP(writer, request)
		if request.URL.Path == "/" {
			gotRequest <- struct{}{}
		}
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
	timeOut := time.Second * 5
	select {
	case <-gotRequest:
		log.Println("got request")
	case <-time.After(timeOut):
		log.Fatalln("timed out")
	}
	select {
	case <-gotPostRequest:
		log.Println("got POST")
	case <-time.After(timeOut):
		log.Fatalln("POST timed out")
	}
	time.Sleep(time.Second * 1)
}
