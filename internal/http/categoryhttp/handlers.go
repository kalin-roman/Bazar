package categoryhttp

import (
	"encoding/json"
	"net/http"

	"github.com/kalin-roman/Bazar/internal/category"
	"github.com/kalin-roman/Bazar/internal/platform/logger"
)

type HandlesService struct {
	CatService *category.Service
}

func NewCategoriesService(s *category.Service) *HandlesService {
	return &HandlesService{CatService: s}
}

func (h *HandlesService) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	category, err := h.CatService.List(ctx)
	if err != nil {
		logger.Error("categoryhttp: List", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category) // writing in JSON

}
