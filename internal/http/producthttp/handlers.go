package producthttp

import (
	"encoding/json"
	"net/http"

	"github.com/kalin-roman/Bazar/internal/product"
)

type HandlesService struct {
	ProductService *product.Service
}

func NewProductService(s *product.Service) *HandlesService {
	return &HandlesService{ProductService: s}
}

func (h *HandlesService) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	products, err := h.ProductService.List(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products) // writing in JSON

}
