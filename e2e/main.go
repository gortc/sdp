package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/runner"
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
	var err error

	// create context
	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create chrome instance
	c, err := chromedp.New(ctxt, chromedp.WithLog(log.Printf), chromedp.WithRunnerOptions(
		runner.Path(*bin),
		runner.DisableGPU,
		runner.Flag("headless", *headless),
	))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		// shutdown chrome
		err = c.Shutdown(ctxt)
		if err != nil {
			log.Fatal(err)
		}

		// wait for chrome to finish
		err = c.Wait()
		if err != nil {
			log.Fatal(err)
		}
	}()
	if err := c.Run(ctxt, chromedp.Navigate(*httpAddr)); err != nil {
		log.Fatalln("failed to navigate:", err)
	}

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
