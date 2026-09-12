package userhttp

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/kalin-roman/Bazar/internal/platform/logger"
	"github.com/kalin-roman/Bazar/internal/platform/middleware"
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
		logger.Error("userhttp: List", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users) // writing in JSON
}

func (h *HandlesService) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Unlike order's GetByID, id here is already a string straight off
	// the path — no strconv.ParseInt needed, since User's own ID is
	// the Supabase UUID directly, not a separate integer PK.
	id := r.PathValue("id")

	u, err := h.UserService.GetByID(ctx, id)
	if errors.Is(err, user.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		logger.Error("userhttp: GetByID", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	requesterID, ok := middleware.UserIDFromContext(ctx)
	if !ok || requesterID != u.ID {
		// Same status as the not-found case above, deliberately — same
		// reasoning as order's GetByID: someone else's profile looks
		// identical to a nonexistent one from the outside. Simpler
		// here than order's check: u.ID *is* the identity being
		// checked, there's no separate owner field to compare against.
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}
