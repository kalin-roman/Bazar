package orderhttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kalin-roman/Bazar/internal/order"
	"github.com/kalin-roman/Bazar/internal/platform/middleware"
	"github.com/kalin-roman/Bazar/internal/product"
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
	o.ID = int64(len(r.orders) + 1)
	o.Status = "Pending"
	r.orders = append(r.orders, o)
	return o, nil
}

// fakeProductRepository is a minimal in-memory product.Repository,
// just enough for Create's server-side price lookup.
type fakeProductRepository struct {
	products []product.Product
}

var _ product.Repository = (*fakeProductRepository)(nil)

func (r *fakeProductRepository) List(ctx context.Context) ([]product.Product, error) {
	return r.products, nil
}

func (r *fakeProductRepository) GetBySlug(ctx context.Context, slug string) (product.Product, error) {
	return product.Product{}, errors.New("not used by this test")
}

func (r *fakeProductRepository) GetByID(ctx context.Context, id int64) (product.Product, error) {
	for _, p := range r.products {
		if p.ID == id {
			return p, nil
		}
	}
	return product.Product{}, product.ErrNotFound
}

func (r *fakeProductRepository) Create(ctx context.Context, p product.Product) (product.Product, error) {
	return product.Product{}, errors.New("not used by this test")
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
			// List never touches ProductService, so a Service backed by
			// an empty fake is enough here.
			h := NewOrderService(svc, product.NewService(&fakeProductRepository{}))

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
	// GetByID never touches ProductService either.
	h := NewOrderService(svc, product.NewService(&fakeProductRepository{}))

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

// TestCreate exercises the checkout endpoint's core guarantee: the
// price actually charged always comes from the product's real,
// current price, never from whatever the client's request body says.
func TestCreate(t *testing.T) {
	orderRepo := &fakeRepository{}
	productRepo := &fakeProductRepository{products: []product.Product{
		{ID: 1, Title: "Chair", PriceCents: 5000},
	}}
	h := NewOrderService(order.NewService(orderRepo), product.NewService(productRepo))

	tests := []struct {
		name       string
		userID     string
		body       string
		wantStatus int
	}{
		{
			name:       "no authenticated user",
			userID:     "",
			body:       `{"items":[{"product_id":1,"quantity":1}]}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "malformed body",
			userID:     "alice",
			body:       `not json`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "product does not exist",
			userID:     "alice",
			body:       `{"items":[{"product_id":999,"quantity":1}]}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "client-supplied price is ignored",
			userID: "alice",
			// A client trying to check out a $50 chair for $1 — the
			// request has no price field at all, so there's nothing
			// for a malicious client to even supply; this just
			// confirms the happy path succeeds and charges the real
			// price server-side.
			body:       `{"items":[{"product_id":1,"quantity":2}]}`,
			wantStatus: http.StatusCreated,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(tc.body))
			if tc.userID != "" {
				r = r.WithContext(middleware.ContextWithUserID(r.Context(), tc.userID))
			}
			w := httptest.NewRecorder()

			h.Create(w, r)

			if w.Code != tc.wantStatus {
				t.Fatalf("got status %d, want %d", w.Code, tc.wantStatus)
			}

			if tc.wantStatus != http.StatusCreated {
				return
			}

			var got order.Order
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("decode response body: %v", err)
			}
			if len(got.Items) != 1 || got.Items[0].PriceCents != 5000 {
				t.Fatalf("got items %+v, want one item priced at 5000 (the product's real price)", got.Items)
			}
		})
	}
}
