package router

import (
	"fmt"
	"net/http"

	"github.com/0xa1-red/empires-of-avalon/actor/admin"
	gamecluster "github.com/0xa1-red/empires-of-avalon/pkg/cluster"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func AdminRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		// r.Use(auth.EnsureValidToken())
		r.Get("/", adminIndex)
	})

	return r
}

func adminIndex(w http.ResponseWriter, r *http.Request) {
	adminGrain := protobuf.GetAdminGrainClient(gamecluster.GetC(), admin.AdminID.String())

	if adminGrain == nil {
		err := fmt.Errorf("failed to retrieve admin grain")
		E(w, r, http.StatusInternalServerError, err)

		return
	}

	res, err := adminGrain.Describe(&protobuf.DescribeAdminRequest{
		TraceID:   "",
		Timestamp: timestamppb.Now(),
	})
	if err != nil {
		err := fmt.Errorf("failed to describe admin grain")
		E(w, r, http.StatusInternalServerError, err)

		return
	}

	render.JSON(w, r, res)
}
