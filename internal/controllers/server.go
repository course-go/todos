package controllers

import (
	"net/http"
	"time"
)

func NewServer(hostname string, mux http.Handler) *http.Server {
	return &http.Server{
		Addr:              hostname,
		ReadTimeout:       1 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       30 * time.Second,
		Handler:           mux,
	}
}
