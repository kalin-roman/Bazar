package userhttp

import (
	"encoding/json"
	"net/http"

	"github.com/kalin-roman/Bazar/internal/user"
)

type HandlesService struct {
	UserService *user.Service
}

func NewUserService(s *user.Service) *HandlesService {
	return &HandlesService{UserService: s}
}

func (h *HandlesService) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, err := h.UserService.List(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users) // writing in JSON
}
