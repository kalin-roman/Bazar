package userhttp

import "net/http"

func RegisterRouter(mux *http.ServeMux, h *HandlesService) {
	mux.HandleFunc("GET /users", h.List)
	mux.HandleFunc("GET /users/{id}", h.GetByID)
}
