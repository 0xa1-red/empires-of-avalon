package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/0xa1-red/empires-of-avalon/api"
	"github.com/0xa1-red/empires-of-avalon/gamecluster"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"golang.org/x/exp/slog"
)

var server *http.Server

func startServer(wg *sync.WaitGroup, addr string) {
	defer wg.Done()
	s := chi.NewRouter()

	s.Use(middleware.Logger)
	s.Use(middleware.RequestID)
	s.Use(middleware.AllowContentType("application/json"))
	s.Use(middleware.Timeout(60 * time.Second))

	s.Mount("/", api.NewRouter(gamecluster.GetC()))

	server = &http.Server{
		Addr:    addr,
		Handler: s,
	}
	slog.Info("starting http server", "address", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("http server error", err)
	}
}
