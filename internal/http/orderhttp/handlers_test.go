package orderhttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kalin-roman/Bazar/internal/order"
	"github.com/kalin-roman/Bazar/internal/platform/middleware"
)

// fakeRepository is a minimal in-memory order.Repository, just enough
// to drive HandlesService's methods directly, without a real database
// or the mux/middleware in front of it.
type fakeRepository struct {
	orders  []order.Order
	listErr error
}

var _ order.Repository = (*fakeRepository)(nil)

func (r *fakeRepository) List(ctx context.Context) ([]order.Order, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.orders, nil
}

func (r *fakeRepository) GetByID(ctx context.Context, id int64) (order.Order, error) {
	for _, o := range r.orders {
		if o.ID == id {
			return o, nil
		}
	}
	return order.Order{}, order.ErrNotFound
}

func (r *fakeRepository) Create(ctx context.Context, o order.Order) (order.Order, error) {
	return order.Order{}, errors.New("not used by this test")
}

func TestList(t *testing.T) {
	tests := []struct {
		name       string
		repo       *fakeRepository
		wantStatus int
	}{
		{
			name: "success",
			repo: &fakeRepository{orders: []order.Order{
				{ID: 1, UserID: "alice"},
			}},
			wantStatus: http.StatusOK,
		},
		{
			name:       "repository error",
			repo:       &fakeRepository{listErr: errors.New("db down")},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := order.NewService(tc.repo)
			h := NewOrderService(svc)

			r := httptest.NewRequest(http.MethodGet, "/orders", nil)
			w := httptest.NewRecorder()

			h.List(w, r)

			if w.Code != tc.wantStatus {
				t.Fatalf("got status %d, want %d", w.Code, tc.wantStatus)
			}
		})
	}
}

// TestGetByID exercises the ownership check built in the
// authorization lesson: an order that exists but belongs to someone
// else must return the exact same status as an order that doesn't
// exist at all (404), so a caller can't tell the two apart.
func TestGetByID(t *testing.T) {
	repo := &fakeRepository{orders: []order.Order{
		{
			ID:     1,
			UserID: "alice",
			Items:  []order.OrderItem{{ProductID: 1, PriceCents: 1000, Quantity: 2}},
		},
	}}
	svc := order.NewService(repo)
	h := NewOrderService(svc)

	tests := []struct {
		name       string
		id         string
		userID     string // "" means no user injected into context at all
		wantStatus int
	}{
		{name: "malformed id", id: "not-a-number", userID: "alice", wantStatus: http.StatusBadRequest},
		{name: "order does not exist", id: "999", userID: "alice", wantStatus: http.StatusNotFound},
		{name: "order belongs to someone else", id: "1", userID: "bob", wantStatus: http.StatusNotFound},
		{name: "no authenticated user in context", id: "1", userID: "", wantStatus: http.StatusNotFound},
		{name: "owner", id: "1", userID: "alice", wantStatus: http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/orders/"+tc.id, nil)
			r.SetPathValue("id", tc.id)
			if tc.userID != "" {
				r = r.WithContext(middleware.ContextWithUserID(r.Context(), tc.userID))
			}
			w := httptest.NewRecorder()

			h.GetByID(w, r)

			if w.Code != tc.wantStatus {
				t.Fatalf("got status %d, want %d", w.Code, tc.wantStatus)
			}

			if tc.wantStatus != http.StatusOK {
				return
			}

			var got order.Order
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("decode response body: %v", err)
			}
			if got.ID != 1 || got.UserID != "alice" {
				t.Fatalf("got order %+v, want order 1 owned by alice", got)
			}
		})
	}
}
