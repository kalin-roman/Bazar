package listinghttp

import (
	"encoding/json"
	"net/http"

	"github.com/kalin-roman/Bazar/internal/listing"
)

type HandlesService struct {
	ListService *listing.Service
}

func NewListingService(s *listing.Service) *HandlesService {
	return &HandlesService{ListService: s}
}

func (h *HandlesService) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	listing, err := h.ListService.List(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listing) // writing in JSON

}
