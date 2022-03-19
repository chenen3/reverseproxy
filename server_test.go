package main

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReverseProxy(t *testing.T) {
	foo := "/foo"
	bar := "1"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Index(r.URL.Path, foo) == 0 {
			logErrorf("pattern not trimmed")
			w.WriteHeader(500)
			return
		}

		if _, err := io.WriteString(w, bar); err != nil {
			logError(err)
			w.WriteHeader(500)
			return
		}
	}))
	defer ts.Close()

	// random address
	ln, _ := net.Listen("tcp", "127.0.0.1:")
	ln.Close()

	conf := &config{
		Listen:    ln.Addr().String(),
		Upstreams: []upstream{{Pattern: foo, Addr: ts.Listener.Addr().String()}},
	}
	srv, err := NewServer(conf, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	r := httptest.NewRequest("GET", "http://127.0.0.1"+foo, nil)
	w := httptest.NewRecorder()
	srv.httpSrv.Handler.ServeHTTP(w, r)

	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("bad status code: %d", resp.StatusCode)
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != bar {
		t.Fatalf("want %s, got %s", bar, bs)
	}
}

// fakeRoundTrip does not execute a real HTTP request,
// it return a response straightly, for better performance profiling
type fakeRoundTrip struct{}

func (fakeRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) {
	_, _ = io.Copy(io.Discard, r.Body)
	_ = r.Body.Close()
	recorder := httptest.NewRecorder()
	_, err := recorder.WriteString("fake response")
	if err != nil {
		return nil, err
	}
	return recorder.Result(), nil
}

func BenchmarkReverseProxyParallel(b *testing.B) {
	conf := &config{
		Listen:    "127.0.0.1:0",
		Upstreams: []upstream{{Pattern: "/foo", Addr: "127.0.0.1:12345"}},
	}
	srv, err := NewServer(conf, fakeRoundTrip{})
	if err != nil {
		b.Fatal(err)
	}
	defer srv.Close()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r := httptest.NewRequest("GET", "http://127.0.0.1/foo", nil)
			w := httptest.NewRecorder()
			srv.httpSrv.Handler.ServeHTTP(w, r)

			resp := w.Result()
			if resp.StatusCode != 200 {
				b.Fatalf("bad status code: %d", resp.StatusCode)
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}
