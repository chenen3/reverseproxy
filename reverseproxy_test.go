package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

// client request (http://localhost:8888) -> reverse proxy -> http testing server (http://localhost:8889)
func TestReverseProxy(t *testing.T) {
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	ln.Close()
	from := ln.Addr().String()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	}))
	defer ts.Close()
	to := ts.URL

	s := http.Server{
		Addr:    from,
		Handler: ReverseProxy(from, to),
	}
	go s.ListenAndServe()
	defer s.Close()

	resp, err := ts.Client().Get("http://" + from)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("bad status code: ", resp.StatusCode)
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != "ok" {
		t.Fatalf("want ok, got %s", bs)
	}
}
