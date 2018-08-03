package main

import (
	"context"
	"flag"
	"fmt"
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
	timeout  = flag.Duration("timeout", time.Second*5, "test timeout")
)

func main() {
	flag.Parse()
	fmt.Println("bin", *bin, "addr", *httpAddr, "timeout", *timeout)
	gotPostRequest := make(chan struct{}, 1)
	fs := http.FileServer(http.Dir("static"))
	http.HandleFunc("/sdp", func(writer http.ResponseWriter, request *http.Request) {
		log.Println("http:", request.Method, request.URL.Path, request.RemoteAddr)
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
	})
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		log.Println("http:", request.Method, request.URL.Path, request.RemoteAddr)
		fs.ServeHTTP(writer, request)
	})
	go func() {
		if err := http.ListenAndServe(*httpAddr, nil); err != nil {
			log.Fatalln("failed to listen:", err)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	c, err := chromedp.New(ctx, chromedp.WithLog(log.Printf), chromedp.WithRunnerOptions(
		runner.Path(*bin), runner.DisableGPU, runner.Flag("headless", *headless),
	))
	if err != nil {
		log.Fatalln("failed to create chrome", err)
	}
	if err := c.Run(ctx, chromedp.Navigate("http://"+*httpAddr)); err != nil {
		log.Fatalln("failed to navigate:", err)
	}
	select {
	case <-gotPostRequest:
		os.Exit(0)
	case <-ctx.Done():
		log.Fatalln("failed to wait until post:", ctx.Err())
	}
}
