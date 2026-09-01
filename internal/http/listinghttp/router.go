package listinghttp

import "net/http"

func NewRouter(h *HandlesService) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /listings", h.List)
	return mux
}
