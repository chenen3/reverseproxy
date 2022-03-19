package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	defaultMaxIdleConnsPerHost = 100
	defaultIdleConnTimeout     = 90 // second
)

type config struct {
	Listen    string     `json:"listen"`
	Upstreams []upstream `json:"upstreams"`

	// optional for optimization
	MaxIdleConnsPerHost int `json:"maxIdleConnsPerHost"`
	IdleConnTimeout     int `json:"idleConnTimeout"` // second
}

type upstream struct {
	Pattern string `json:"pattern"`
	Addr    string `json:"addr"`
}

// Server provides reverse proxy service
type Server struct {
	httpSrv *http.Server
	rt      http.RoundTripper
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewServer initialize Server with specified config and RoundTripper,
// if the RoundTripper is nil, then customized http.Transport will be used
func NewServer(c *config, rt http.RoundTripper) (*Server, error) {
	if c.MaxIdleConnsPerHost == 0 {
		c.MaxIdleConnsPerHost = defaultMaxIdleConnsPerHost
	}
	if c.IdleConnTimeout == 0 {
		c.IdleConnTimeout = defaultIdleConnTimeout
	}
	ctx, cancel := context.WithCancel(context.Background())
	s := Server{
		ctx:    ctx,
		cancel: cancel,
	}

	if rt == nil {
		rt = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2: true,
			// MaxIdleConns:          100,
			MaxIdleConnsPerHost:   c.MaxIdleConnsPerHost,
			IdleConnTimeout:       time.Duration(c.IdleConnTimeout) * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	}
	s.rt = rt

	mux := new(http.ServeMux)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if verbose {
			logInfof("peer %s, path %q, upstream %s", r.RemoteAddr, r.URL.Path, "default")
		}
		_, err := io.WriteString(w, "welcome\n")
		if err != nil {
			w.WriteHeader(500)
			logError(err)
		}
	})
	for _, upstream := range c.Upstreams {
		if upstream.Pattern == "" || upstream.Addr == "" {
			return nil, fmt.Errorf("invalid upstream: %v", upstream)
		}
		mux.HandleFunc(upstream.Pattern, func(w http.ResponseWriter, r *http.Request) {
			if verbose {
				logInfof("peer %s, path %q, upstream %s", r.RemoteAddr, r.URL.Path, upstream.Addr)
			}
			u := *r.URL
			u.Scheme = "http"
			u.Host = upstream.Addr
			u.Path = strings.TrimPrefix(u.Path, upstream.Pattern)
			ctxR, cancelR := context.WithTimeout(s.ctx, time.Minute)
			defer cancelR()
			req, err := http.NewRequestWithContext(ctxR, r.Method, u.String(), r.Body)
			if err != nil {
				logError(err)
				w.WriteHeader(500)
				return
			}
			req.RemoteAddr = r.RemoteAddr

			resp, err := s.rt.RoundTrip(req)
			if err != nil {
				logError(err)
				w.WriteHeader(500)
				return
			}
			defer resp.Body.Close()
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				logError(err)
				w.WriteHeader(500)
				return
			}
		})
	}
	s.httpSrv = &http.Server{
		Addr:    c.Listen,
		Handler: mux,
	}
	return &s, nil
}

func (s *Server) Serve() error {
	return s.httpSrv.ListenAndServe()
}

func (s *Server) Close() error {
	s.cancel()
	return s.httpSrv.Close()
}
