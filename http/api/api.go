package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/0xa1-red/empires-of-avalon/common"
	"github.com/0xa1-red/empires-of-avalon/http/middleware"
	"github.com/0xa1-red/empires-of-avalon/instrumentation/traces"
	"github.com/0xa1-red/empires-of-avalon/pkg/auth"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/go-chi/chi"
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

func NewRouter(c *cluster.Cluster) *Router {
	r := chi.NewRouter()

	router := &Router{
		Router:  r,
		cluster: c,
	}

	r.Get("/", router.Index)

	r.Group(func(r chi.Router) {
		r.Use(auth.IsAuthenticated)
		r.Get("/inventory", router.Inventory)
		r.Post("/build", router.Build)
	})

	return router
}

func (rt *Router) Index(w http.ResponseWriter, r *http.Request) {
	resp := CommonResponse{
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

	auth := authFromContext(ctx)

	authUUID, err := uuid.Parse(auth)
	if err != nil {
		span.RecordError(err)
		slog.Error("failed to parse authorization header", err, "auth", auth, "url", r.URL.String())

		status := http.StatusBadRequest
		res := ErrorResponse{
			Status:     status,
			StatusText: http.StatusText(status),
			Error:      err,
		}

		render.Status(r, status)
		render.JSON(w, r, res)

		return
	}

	span.SetAttributes(attribute.String("user_id", authUUID.String()))
	slog.Info("getting inventory grain client", "id", common.GetInventoryID(authUUID).String())

	inventory := protobuf.GetInventoryGrainClient(rt.cluster, common.GetInventoryID(authUUID).String())

	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, &carrier)

	res, err := inventory.Describe(&protobuf.DescribeInventoryRequest{
		TraceID:   carrier.Get("traceparent"),
		Timestamp: timestamppb.Now(),
	})
	if err != nil {
		span.RecordError(err)
		slog.Error("failed to get inventory", err, "auth", auth, "url", r.URL.String())

		status := http.StatusInternalServerError
		res := ErrorResponse{
			Status:     status,
			StatusText: http.StatusText(status),
			Error:      err,
		}

		render.Status(r, status)
		render.JSON(w, r, res)

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

	auth := authFromContext(ctx)

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var buildRequest BuildRequest
	if err := decoder.Decode(&buildRequest); err != nil {
		span.RecordError(err)
		slog.Error("failed to parse request", err,
			"auth", auth,
		)

		status := http.StatusBadRequest
		res := ErrorResponse{
			Status:     status,
			StatusText: http.StatusText(status),
			Error:      err,
		}

		render.Status(r, status)
		render.JSON(w, r, res)

		return
	}

	building := buildRequest.Building

	span.SetAttributes(attribute.String("building", building))

	amt := getBuildingAmount(buildRequest)

	b, ok := common.Buildings[common.BuildingName(building)]
	if !ok {
		err := fmt.Errorf("invalid building type %s", building)
		span.RecordError(err)
		slog.Error("failed to start building", err,
			"auth", auth,
			"building", building,
		)

		status := http.StatusNotFound
		res := ErrorResponse{
			Status:     status,
			StatusText: http.StatusText(status),
			Error:      err,
		}

		render.Status(r, status)
		render.JSON(w, r, res)

		return
	}

	authUUID, err := uuid.Parse(auth)
	if err != nil {
		span.RecordError(err)
		slog.Error("failed to parse authorization header", err, "auth", auth, "url", r.URL.String())

		status := http.StatusBadRequest
		res := ErrorResponse{
			Status:     status,
			StatusText: http.StatusText(status),
			Error:      err,
		}

		render.Status(r, status)
		render.JSON(w, r, res)

		return
	}

	span.SetAttributes(attribute.String("user_id", authUUID.String()))
	inventory := protobuf.GetInventoryGrainClient(rt.cluster, common.GetInventoryID(authUUID).String())

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

		status := http.StatusInternalServerError
		res := ErrorResponse{
			Status:     status,
			StatusText: http.StatusText(status),
			Error:      err,
		}

		render.Status(r, status)
		render.JSON(w, r, res)

		return
	}

	if res.Status == protobuf.Status_Error {
		span.RecordError(err)
		slog.Error("failed to start building", fmt.Errorf("%s", res.Error),
			"auth", auth,
			"url", r.URL.String(),
			"building", b.Name,
		)

		status := http.StatusInternalServerError
		res := ErrorResponse{
			Status:     status,
			StatusText: http.StatusText(status),
			Error:      err,
		}

		render.Status(r, status)
		render.JSON(w, r, res)

		return
	}

	status := http.StatusCreated
	resp := CommonResponse{
		Status:     status,
		StatusText: http.StatusText(status),
		Message:    "OK",
	}
	render.JSON(w, r, resp)
}

func getBuildingAmount(r BuildRequest) int64 {
	if !queueBuildings {
		return 1
	}

	amt := int64(r.Amount)

	if amt > maximumBuildingRequest {
		return maximumBuildingRequest
	}

	return amt
}

func E(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("unauthorized", err, "url", r.URL.String())

	status := http.StatusUnauthorized
	res := ErrorResponse{
		Status:     status,
		StatusText: http.StatusText(status),
		Error:      err,
	}

	render.Status(r, status)
	render.JSON(w, r, res)
	return
}

func authFromContext(ctx context.Context) string {
	userProfile := ctx.Value(auth.ContextUserProfile)
	if userProfile == nil {
		return ""
	}

	var auth string

	if p, ok := userProfile.(map[string]interface{}); ok {
		if appMetadata, ok := p["app_metadata"].(map[string]interface{}); ok {
			if authRaw, ok := appMetadata["external_id"].(string); ok {
				auth = authRaw
			}
		}
	}

	return auth
}
