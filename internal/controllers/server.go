package controllers

import (
	"net/http"
	"time"
)

const (
	defaultServerReadHeaderTimeout  = 2 * time.Second
	defaultServerIdleTimeoutTimeout = 30 * time.Second
)

func NewServer(hostname string, mux http.Handler) *http.Server {
	return &http.Server{
		Addr:              hostname,
		ReadTimeout:       1 * time.Second,
		ReadHeaderTimeout: defaultServerReadHeaderTimeout,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       defaultServerIdleTimeoutTimeout,
		Handler:           mux,
	}
}
