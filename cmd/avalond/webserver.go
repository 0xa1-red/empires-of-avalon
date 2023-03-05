package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/0xa1-red/empires-of-avalon/common"
	"github.com/0xa1-red/empires-of-avalon/gamecluster"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var server *http.Server

func startServer(wg *sync.WaitGroup, addr string) {
	defer wg.Done()
	s := mux.NewRouter()

	s.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

	s.HandleFunc("/inventory", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")

		c := gamecluster.GetC()

		authUUID, err := uuid.Parse(auth)
		if err != nil {
			slog.Error("failed to parse authorization header", err, "auth", auth, "url", r.URL.String())
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		inventory := protobuf.GetInventoryGrainClient(c, common.GetInventoryID(authUUID).String())

		res, err := inventory.Describe(&protobuf.DescribeInventoryRequest{})
		if err != nil {
			slog.Error("failed to get inventory", err, "auth", auth, "url", r.URL.String())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		raw, err := res.Inventory.MarshalJSON()
		if err != nil {
			slog.Error("failed to marshal response", err, "auth", auth, "url", r.URL.String())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(raw)
	})

	s.HandleFunc("/build", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		building := r.URL.Query().Get("building")

		amount := r.URL.Query().Get("amount")
		amt, err := strconv.ParseInt(amount, 10, 64)
		if err != nil {
			slog.Error("failed to parse amount", err)
			amt = 1
		}

		b, ok := common.Buildings[common.BuildingName(building)]
		if !ok {
			slog.Error("failed to start building", fmt.Errorf("invalid building type %s", building),
				"auth", auth,
				"building", building,
			)
			http.Error(w, fmt.Sprintf("invalid building: %s", building), http.StatusNotFound)
			return
		}

		c := gamecluster.GetC()

		authUUID, err := uuid.Parse(auth)
		if err != nil {
			slog.Error("failed to parse authorization header", err, "auth", auth, "url", r.URL.String())
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		inventory := protobuf.GetInventoryGrainClient(c, common.GetInventoryID(authUUID).String())

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
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if res.Status == protobuf.Status_Error {
			slog.Error("failed to start building", fmt.Errorf("%s", res.Error),
				"auth", auth,
				"url", r.URL.String(),
				"building", b.Name,
			)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Write([]byte("OK"))
	})

	server = &http.Server{
		Addr:    addr,
		Handler: s,
	}
	slog.Info("starting http server", "address", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("http server error", err)
	}
}
