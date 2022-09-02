package main

import (
	"syscall"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	var from, to string
	flag.StringVar(&from, "from", "", "listen address")
	flag.StringVar(&to, "to", "", "the URL that proxy to")
	flag.Parse()
	if from == "" {
		fmt.Println("option -from required")
		return
	}
	if to == "" {
		fmt.Println("option -to required")
		return
	}

	s := http.Server{
		Addr:    from,
		Handler: ReverseProxy(from, to),
	}
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, os.Interrupt)
		<-ch
		s.Close()
	}()
	err := s.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
