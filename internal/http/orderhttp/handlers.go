package orderhttp

import (
	"encoding/json"
	"net/http"

	"github.com/kalin-roman/Bazar/internal/order"
)

type HandlesService struct {
	OrderService *order.Service
}

func NewOrderService(s *order.Service) *HandlesService {
	return &HandlesService{OrderService: s}
}

func (h *HandlesService) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orders, err := h.OrderService.List(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders) // writing in JSON
}
