package orderhttp

import "net/http"

func RegisterRouter(mux *http.ServeMux, h *HandlesService) {
	mux.HandleFunc("GET /orders", h.List)
}
