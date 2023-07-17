package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/0xa1-red/empires-of-avalon/api"
	"github.com/0xa1-red/empires-of-avalon/gamecluster"
	intmw "github.com/0xa1-red/empires-of-avalon/http/middleware"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"golang.org/x/exp/slog"
)

var server *http.Server

func startServer(wg *sync.WaitGroup, addr string) {
	defer wg.Done()

	s := chi.NewRouter()

	s.Use(middleware.Logger)
	s.Use(intmw.AvalonLogger)
	s.Use(middleware.AllowContentType("application/json"))
	s.Use(middleware.Timeout(60 * time.Second))

	s.Mount("/", api.NewRouter(gamecluster.GetC()))

	server = &http.Server{ // nolint:exhaustruct
		Addr:              addr,
		Handler:           s,
		ReadHeaderTimeout: 3 * time.Second,
	}

	slog.Info("starting http server", "address", addr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("http server error", err)
	}
}
