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
		logError(err)
		return
	}
	var c config
	err = json.NewDecoder(f).Decode(&c)
	if err != nil {
		logError(err)
		return
	}

	setVerbose(true)

	s, err := NewServer(&c, nil)
	if err != nil {
		logError(err)
		return
	}
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		if e := s.Close(); e != nil {
			panic(err)
		}
	}()
	logInfof("start server")
	err = s.Serve()
	if err != nil && err != http.ErrServerClosed {
		logError(err)
		return
	}
}
