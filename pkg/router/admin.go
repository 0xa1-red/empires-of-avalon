package router

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/0xa1-red/empires-of-avalon/actor/admin"
	"github.com/0xa1-red/empires-of-avalon/pkg/assets"
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
		r.Get("/dashboard", adminIndex)
	})

	return r
}

func adminIndex(w http.ResponseWriter, r *http.Request) {
	t, err := assets.LoadTemplate("/assets/templates/admin/index.gohtml")
	if err != nil {
		E(w, r, http.StatusInternalServerError, err)
		return
	}

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

	data := struct {
		Timestamp time.Time
		Data      map[string]any
	}{
		Timestamp: res.Timestamp.AsTime(),
		Data:      res.Admin.AsMap(),
	}

	HTML(w, r, t, data)
}

func HTML(w http.ResponseWriter, r *http.Request, t *template.Template, data any) {
	buf := bytes.NewBuffer([]byte(""))
	if err := t.Execute(buf, data); err != nil {
		E(w, r, http.StatusInternalServerError, err)
		return
	}

	render.HTML(w, r, buf.String())
}
