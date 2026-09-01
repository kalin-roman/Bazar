package categoryhttp

import "net/http"

func NewRouter(h *HandlesService) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /categories", h.List)
	return mux

}
