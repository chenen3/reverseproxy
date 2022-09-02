package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Handler struct {
	from string // localhost:8888
	to   string // http://localhost:8889/
}

// ReverseProxy return a http.Handler that customized for reverse proxy
func ReverseProxy(from, to string) *Handler {
	return &Handler{from: from, to: to}
}

var dialer = &net.Dialer{
	Timeout:   30 * time.Second,
	KeepAlive: 30 * time.Second,
}

var transport http.RoundTripper = &http.Transport{
	DialContext:           dialer.DialContext,
	ForceAttemptHTTP2:     true,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	toURL, err := url.Parse(h.to)
	if err != nil {
		log.Print("parse toURL: ", err)
		w.WriteHeader(500)
		return
	}

	req, err := http.NewRequest(r.Method, toURL.String(), r.Body)
	if err != nil {
		log.Print("make request: ", err)
		w.WriteHeader(500)
		return
	}
	req.Header = r.Header.Clone()
	req.Header.Set("Host", toURL.Hostname())
	// do not use http Client, it will handle redirect by default
	resp, err := transport.RoundTrip(req)
	if err != nil {
		log.Print("forward request: ", err)
		w.WriteHeader(500)
		return
	}
	defer resp.Body.Close()
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
