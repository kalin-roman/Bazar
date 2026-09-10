package producthttp

import "net/http"

func RegisterRouter(mux *http.ServeMux, h *HandlesService) {
	mux.HandleFunc("GET /products", h.List)
}
