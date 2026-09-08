package listinghttp

import "net/http"

func RegisterRouter(mux *http.ServeMux, h *HandlesService) {
	mux.HandleFunc("GET /listings", h.List)
}
