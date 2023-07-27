package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/0xa1-red/empires-of-avalon/instrumentation/traces"
	"github.com/0xa1-red/empires-of-avalon/pkg/middleware"
	"github.com/0xa1-red/empires-of-avalon/pkg/model"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/auth"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/blueprints"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/game"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/registry"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Router struct {
	chi.Router

	cluster *cluster.Cluster
}

func NewGameRouter(c *cluster.Cluster) *Router {
	r := chi.NewRouter()

	router := &Router{
		Router:  r,
		cluster: c,
	}

	r.Use(cors.Handler(cors.Options{ // nolint:exhaustruct
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization"},
		MaxAge:           300,
		Debug:            true,
	}))

	r.Group(func(r chi.Router) {
		r.Use(auth.EnsureValidToken())
		r.Get("/inventory", router.Inventory)
		r.Post("/build", router.Build)
	})

	return router
}

func (rt *Router) Index(w http.ResponseWriter, r *http.Request) {
	resp := model.CommonResponse{
		Status:     http.StatusOK,
		StatusText: http.StatusText(http.StatusOK),
		Message:    "Hi",
	}

	render.JSON(w, r, resp)
}

func (rt *Router) Inventory(w http.ResponseWriter, r *http.Request) {
	ctx, span := traces.Start(r.Context(), "api/router/inventory")
	defer span.End()

	ctx = context.WithValue(ctx, middleware.RequestIDKey, span.SpanContext().TraceID())
	r = r.WithContext(ctx)

	w.Header().Set("X-Trace-Id", span.SpanContext().TraceID().String())

	auth := authFromContext(w, r, ctx)

	authUUID, err := uuid.Parse(auth)
	if err != nil {
		span.RecordError(err)
		slog.Error("failed to parse authorization header", err, "auth", auth, "url", r.URL.String())
		E(w, r, http.StatusBadRequest, err)

		return
	}

	span.SetAttributes(attribute.String("user_id", authUUID.String()))

	res, err := game.Describe(ctx, authUUID)

	if err != nil {
		span.RecordError(err)
		slog.Error("failed to get inventory", err, "auth", auth, "url", r.URL.String())
		E(w, r, http.StatusInternalServerError, err)

		return
	}

	render.JSON(w, r, res)
}

func (rt *Router) Build(w http.ResponseWriter, r *http.Request) {
	ctx, span := traces.Start(r.Context(), "api/router/build")
	defer span.End()

	ctx = context.WithValue(ctx, middleware.RequestIDKey, span.SpanContext().TraceID())
	r = r.WithContext(ctx)

	w.Header().Set("X-Trace-Id", span.SpanContext().TraceID().String())

	auth := authFromContext(w, r, ctx)

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var buildRequest model.BuildRequest
	if err := decoder.Decode(&buildRequest); err != nil {
		span.RecordError(err)
		slog.Error("failed to parse request", err,
			"auth", auth,
		)
		E(w, r, http.StatusBadRequest, err)

		return
	}

	span.SetAttributes(attribute.String("building", buildRequest.Building))

	building := buildRequest.Building

	amt := game.GetBuildingAmount(buildRequest)

	buildingID := blueprints.GetBuildingID(building)

	b, err := registry.GetBuilding(buildingID)
	if err != nil {
		span.RecordError(err)
		slog.Error("failed to start building", err,
			"auth", auth,
			"building", building,
		)
		E(w, r, http.StatusNotFound, err)

		return
	}

	authUUID, err := uuid.Parse(auth)
	if err != nil {
		span.RecordError(err)
		slog.Error("failed to parse authorization header", err, "auth", auth, "url", r.URL.String())
		E(w, r, http.StatusBadRequest, err)

		return
	}

	span.SetAttributes(attribute.String("user_id", authUUID.String()))
	inventory := protobuf.GetInventoryGrainClient(rt.cluster, blueprints.GetInventoryID(authUUID).String())

	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, &carrier)

	res, err := inventory.Start(&protobuf.StartRequest{
		TraceID:   carrier.Get("traceparent"),
		Name:      string(b.Name),
		Amount:    amt,
		Timestamp: timestamppb.Now(),
	})

	if err != nil {
		span.RecordError(err)
		slog.Error("failed to start building", err,
			"auth", auth,
			"url", r.URL.String(),
			"building", b.Name,
		)
		E(w, r, http.StatusInternalServerError, err)

		return
	}

	if res.Status == protobuf.Status_Error {
		span.RecordError(err)
		slog.Error("failed to start building", fmt.Errorf("%s", res.Error),
			"auth", auth,
			"url", r.URL.String(),
			"building", b.Name,
		)
		E(w, r, http.StatusInternalServerError, err)

		return
	}

	status := http.StatusCreated
	resp := model.CommonResponse{
		Status:     status,
		StatusText: http.StatusText(status),
		Message:    "OK",
	}
	render.JSON(w, r, resp)
}
