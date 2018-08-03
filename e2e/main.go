package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/runner"
	"github.com/gortc/sdp"
)

var (
	bin      = flag.String("b", "/usr/bin/google-chrome", "path to binary")
	headless = flag.Bool("headless", true, "headless mode")
	httpAddr = flag.String("addr", "127.0.0.1:5568", "http endpoint to listen")
)

func main() {
	flag.Parse()
	gotRequest := make(chan struct{}, 1)
	gotPostRequest := make(chan struct{}, 1)
	fs := http.FileServer(http.Dir("static"))
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		log.Println("http:", request.Method, request.URL.Path, request.RemoteAddr)
		if request.Method == http.MethodPost {
			log.Println("GOT POST")
			buf, err := ioutil.ReadAll(request.Body)
			if err != nil {
				log.Fatalln("failed to read:", err)
			}
			s, err := sdp.DecodeSession(buf, nil)
			if err != nil {
				log.Fatalln("failed to decode session:", err)
			}
			decoder := sdp.NewDecoder(s)
			message := new(sdp.Message)
			if err := decoder.Decode(message); err != nil {
				log.Fatalln("failed to decode message:", err)
			}
			log.Println("decoded address:", message.Origin.Address)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// create chrome instance
	c, err := chromedp.New(ctx, chromedp.WithLog(log.Printf), chromedp.WithRunnerOptions(
		runner.Path(*bin),
		runner.DisableGPU,
		runner.Flag("headless", *headless),
	))
	if err != nil {
		log.Fatal(err)
	}
	if err := c.Run(ctx, chromedp.Navigate("http://"+*httpAddr)); err != nil {
		log.Fatalln("failed to navigate:", err)
	}
	if err := c.Run(ctx, chromedp.WaitVisible(`#title`, chromedp.ByID)); err != nil {
		log.Fatalln("failed to wait:", err)
	}
	select {
	case <-gotRequest:
		log.Println("got request")
	case <-ctx.Done():
		log.Fatalln("request waiting out")
	}
	select {
	case <-gotPostRequest:
		log.Println("got POST")
		os.Exit(0)
	case <-ctx.Done():
		log.Fatalln("POST timed out")
	}
	log.Println("END")
}
