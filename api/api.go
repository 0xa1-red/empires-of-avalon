package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/0xa1-red/empires-of-avalon/common"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/google/uuid"
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
	r.Get("/inventory", router.Inventory)
	r.Post("/build", router.Build)

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
	auth := r.Header.Get("Authorization")

	authUUID, err := uuid.Parse(auth)
	if err != nil {
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
	inventory := protobuf.GetInventoryGrainClient(rt.cluster, common.GetInventoryID(authUUID).String())

	res, err := inventory.Describe(&protobuf.DescribeInventoryRequest{})
	if err != nil {
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
	auth := r.Header.Get("Authorization")

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var buildRequest BuildRequest
	if err := decoder.Decode(&buildRequest); err != nil {
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

	amt := getBuildingAmount(buildRequest)

	b, ok := common.Buildings[common.BuildingName(building)]
	if !ok {
		err := fmt.Errorf("invalid building type %s", building)
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
	inventory := protobuf.GetInventoryGrainClient(rt.cluster, common.GetInventoryID(authUUID).String())

	res, err := inventory.Start(&protobuf.StartRequest{
		Name:      string(b.Name),
		Amount:    amt,
		Timestamp: timestamppb.Now(),
	})

	if err != nil {
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
