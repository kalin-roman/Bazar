package userhttp

import "net/http"

func NewRouter(h *HandlesService) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users", h.List)
	return mux
}
