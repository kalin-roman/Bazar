package orderhttp

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/kalin-roman/Bazar/internal/order"
	"github.com/kalin-roman/Bazar/internal/platform/logger"
	"github.com/kalin-roman/Bazar/internal/platform/middleware"
	"github.com/kalin-roman/Bazar/internal/product"
)

type HandlesService struct {
	OrderService *order.Service
	// Used only to look up each item's *current* price server-side
	// during Create — a checkout request supplies product IDs and
	// quantities, never a price, so a client can't dictate what it
	// pays.
	ProductService *product.Service
}

func NewOrderService(s *order.Service, p *product.Service) *HandlesService {
	return &HandlesService{OrderService: s, ProductService: p}
}

func (h *HandlesService) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orders, err := h.OrderService.List(ctx)
	if err != nil {
		logger.Error("orderhttp: List", err)
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
		logger.Error("orderhttp: GetByID", err) // some other, unexpected failure
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

// createOrderItem is what a checkout request actually sends per line
// item — a product and how many, nothing else. No price field: the
// client doesn't get to say what anything costs.
type createOrderItem struct {
	ProductID int64 `json:"product_id"`
	Quantity  int64 `json:"quantity"`
}

type createOrderRequest struct {
	Items []createOrderItem `json:"items"`
}

func (h *HandlesService) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Auth already ran and verified the token before this handler was
	// ever reached, so this should always be present — checked anyway
	// rather than assumed, same defensive shape as GetByID's ownership
	// check.
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req createOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	items := make([]order.OrderItem, 0, len(req.Items))
	for _, reqItem := range req.Items {
		p, err := h.ProductService.GetByID(ctx, reqItem.ProductID)
		if errors.Is(err, product.ErrNotFound) {
			// A product ID that doesn't exist is a bad request, not a
			// server failure — the client asked to buy something that
			// isn't there.
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err != nil {
			logger.Error("orderhttp: Create: lookup product", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		items = append(items, order.OrderItem{
			ProductID: reqItem.ProductID,
			// The authoritative price: whatever the product's real,
			// current price is right now, not anything the client sent.
			PriceCents: p.PriceCents,
			Quantity:   reqItem.Quantity,
		})
	}

	created, err := h.OrderService.Create(ctx, order.Order{UserID: userID, Items: items})
	if errors.Is(err, order.ErrInvalid) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err != nil {
		logger.Error("orderhttp: Create", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}
