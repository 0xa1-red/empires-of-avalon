package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/0xa1-red/empires-of-avalon/pkg/router"
	"golang.org/x/exp/slog"
)

var server *http.Server

func startServer(wg *sync.WaitGroup, addr string) {
	defer wg.Done()

	s := router.New()

	server = &http.Server{
		Addr:              addr,
		Handler:           s,
		ReadHeaderTimeout: 3 * time.Second,
	}

	slog.Info("starting http server", "address", addr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("http server error", err)
	}
}
