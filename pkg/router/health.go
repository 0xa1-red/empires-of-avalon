package router

import (
	"net/http"
	"time"

	"github.com/go-chi/render"
)

func Healthcheck(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{
		"status":    "OK",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
