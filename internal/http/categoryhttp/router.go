package categoryhttp

import "net/http"

func RegisterRouter(mux *http.ServeMux, h *HandlesService) {
	mux.HandleFunc("GET /categories", h.List)
}
