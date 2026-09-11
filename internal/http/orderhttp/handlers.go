package orderhttp

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/kalin-roman/Bazar/internal/http/middleware"
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
		log.Println("orderhttp: List:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders) // writing in JSON
}

func (h *HandlesService) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	valueID := r.PathValue("id")
	parseInInt, errPars := strconv.ParseInt(valueID, 10, 64)
	if errPars != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	o, err := h.OrderService.GetByID(ctx, parseInInt)
	if errors.Is(err, order.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println("orderhttp: GetByID:", err) // some other, unexpected failure
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok || userID != o.UserID {
		// Same status as the not-found case above, deliberately — an
		// order that exists but isn't yours looks identical to one
		// that doesn't exist at all, so ownership can't be probed.
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(o)
}
