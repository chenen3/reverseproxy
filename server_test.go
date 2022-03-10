package main

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var (
	foo = "/foo"
	bar = "1"
	ts  *httptest.Server
	srv *Server
)

func TestMain(m *testing.M) {
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Index(r.URL.Path, foo) == 0 {
			logger.Errorf("pattern not trimmed")
			w.WriteHeader(500)
			return
		}

		if _, err := io.WriteString(w, bar); err != nil {
			logger.Error(err)
			w.WriteHeader(500)
			return
		}
	}))
	defer ts.Close()

	// random address
	ln, _ := net.Listen("tcp", "127.0.0.1:")
	ln.Close()

	var err error
	srv, err = NewServer(&config{
		Listen: ln.Addr().String(),
		Upstreams: []upstream{
			{Pattern: foo, Addr: ts.Listener.Addr().String()},
		},
	})
	if err != nil {
		logger.Fatal(err)
	}
	defer srv.Close()
	go func() {
		if e := srv.Serve(); e != nil && e != http.ErrServerClosed {
			logger.Error(e)
		}
	}()
	<-srv.ready

	os.Exit(m.Run())
}

func TestServer(t *testing.T) {
	resp, err := http.Get("http://" + srv.httpSrv.Addr + foo)
	if err != nil {
		t.Fatal(err)
	}
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

func TestReverseProxy(t *testing.T) {
	// the request go straight to handler but not server,
	// so request's target does not matter here
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

func BenchmarkReverseProxyParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r := httptest.NewRequest("GET", "http://127.0.0.1"+foo, nil)
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
