package main

import (
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	f, err := os.Open("config.json")
	if err != nil {
		logger.Fatal(err)
	}
	var c config
	err = json.NewDecoder(f).Decode(&c)
	if err != nil {
		logger.Fatal(err)
	}

	s, err := NewServer(&c)
	if err != nil {
		logger.Fatal(err)
	}
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		if e := s.Close(); e != nil {
			logger.Fatal(e)
		}
	}()
	err = s.Serve()
	if err != nil && err != http.ErrServerClosed {
		logger.Fatal(err)
	}
}
