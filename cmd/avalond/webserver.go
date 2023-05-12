package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/0xa1-red/empires-of-avalon/api"
	"github.com/0xa1-red/empires-of-avalon/gamecluster"
	"github.com/0xa1-red/empires-of-avalon/graph"
	"github.com/0xa1-red/empires-of-avalon/httpserver"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"golang.org/x/exp/slog"
)

var server *http.Server

func startServer(wg *sync.WaitGroup, addr string) {
	defer wg.Done()
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	s := chi.NewRouter()

	s.Use(middleware.Logger)
	s.Use(middleware.RequestID)
	s.Use(middleware.AllowContentType("application/json"))
	s.Use(middleware.Timeout(60 * time.Second))

	s.Handle("/playground", playground.Handler("GraphQL playground", "/graph/query"))
	s.Route("/graph", func(r chi.Router) {
		r.Use(httpserver.Authentication)
		r.Handle("/query", srv)
	})
	s.Mount("/api", api.NewRouter(gamecluster.GetC()))

	server = &http.Server{
		Addr:    addr,
		Handler: s,
	}
	slog.Info("starting http server", "address", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("http server error", err)
	}
}
