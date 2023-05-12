package httpserver

import (
	"net/http"
)

func E(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}
