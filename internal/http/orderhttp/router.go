package orderhttp

import "net/http"

func NewRouter(h *HandlesService) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /orders", h.List)
	return mux
}
