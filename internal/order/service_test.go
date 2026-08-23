package order

import (
	"context"
	"errors"
	"testing"
)

// fakeRepository is an in-memory Repository, used so Service can be
// tested without a real database.
type fakeRepository struct {
	orders []Order
}

var _ Repository = (*fakeRepository)(nil)

var errNotFound = errors.New("order: not found")

func (r *fakeRepository) List(ctx context.Context) ([]Order, error) {
	return r.orders, nil
}

func (r *fakeRepository) GetByID(ctx context.Context, id int64) (Order, error) {
	for _, o := range r.orders {
		if o.ID == id {
			return o, nil
		}
	}
	return Order{}, errNotFound
}

func (r *fakeRepository) Create(ctx context.Context, o Order) (Order, error) {
	r.orders = append(r.orders, o)
	return o, nil
}

func TestServiceList(t *testing.T) {
	repo := &fakeRepository{orders: []Order{
		{ID: 1, UserID: 1},
		{ID: 2, UserID: 2},
	}}
	s := NewService(repo)

	got, err := s.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("List returned %d orders, want 2", len(got))
	}
}

func TestServiceGetByID(t *testing.T) {
	repo := &fakeRepository{orders: []Order{
		{ID: 1, UserID: 1},
	}}
	s := NewService(repo)

	tests := []struct {
		name       string
		id         int64
		wantUserID int64
		wantErr    error
	}{
		{name: "found", id: 1, wantUserID: 1},
		{name: "missing", id: 99, wantErr: errNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := s.GetByID(context.Background(), tc.id)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("got err %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetByID returned error: %v", err)
			}
			if got.UserID != tc.wantUserID {
				t.Fatalf("got UserID %d, want %d", got.UserID, tc.wantUserID)
			}
		})
	}
}

func TestServiceCreate(t *testing.T) {
	valid := Order{
		UserID: 1,
		Items: []OrderItem{
			{ListingID: 1, Price: 1000, Quantity: 2},
		},
	}

	tests := []struct {
		name    string
		mutate  func(o Order) Order
		wantErr error
	}{
		{
			name:   "valid",
			mutate: func(o Order) Order { return o },
		},
		{
			name:    "no user",
			mutate:  func(o Order) Order { o.UserID = 0; return o },
			wantErr: ErrInvalid,
		},
		{
			name:    "no items",
			mutate:  func(o Order) Order { o.Items = nil; return o },
			wantErr: ErrInvalid,
		},
		{
			name: "item with no listing",
			mutate: func(o Order) Order {
				o.Items = []OrderItem{{ListingID: 0, Price: 1000, Quantity: 1}}
				return o
			},
			wantErr: ErrInvalid,
		},
		{
			name: "item with zero quantity",
			mutate: func(o Order) Order {
				o.Items = []OrderItem{{ListingID: 1, Price: 1000, Quantity: 0}}
				return o
			},
			wantErr: ErrInvalid,
		},
		{
			name: "item with negative quantity",
			mutate: func(o Order) Order {
				o.Items = []OrderItem{{ListingID: 1, Price: 1000, Quantity: -1}}
				return o
			},
			wantErr: ErrInvalid,
		},
		{
			name: "item with negative price",
			mutate: func(o Order) Order {
				o.Items = []OrderItem{{ListingID: 1, Price: -1, Quantity: 1}}
				return o
			},
			wantErr: ErrInvalid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepository{}
			s := NewService(repo)

			got, err := s.Create(context.Background(), tc.mutate(valid))

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("got err %v, want %v", err, tc.wantErr)
				}
				if len(repo.orders) != 0 {
					t.Fatalf("invalid Create should not have reached the repository")
				}
				return
			}
			if err != nil {
				t.Fatalf("Create returned error: %v", err)
			}
			if got.UserID != valid.UserID {
				t.Fatalf("got UserID %d, want %d", got.UserID, valid.UserID)
			}
			if len(repo.orders) != 1 {
				t.Fatalf("expected Create to store the order in the repository")
			}
		})
	}
}
