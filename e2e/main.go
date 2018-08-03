package main

import "github.com/mkenney/go-chrome/tot"

func main() {
	f := chrome.Flags{}
	f.Set("headless", "true")
	f.Set("disable-gpu", "true")
	f.Set("disable-dev-shm-usage", "true")
	c := chrome.New(f, "/usr/bin/google-chrome-unstable", "", "", "")
	if err := c.Launch(); err != nil {
		panic(err)
	}
	if err := c.Close(); err != nil {
		panic(err)
	}
}
