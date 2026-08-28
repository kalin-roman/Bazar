package order

import (
	"context"
	"errors"
)

var ErrInvalid = errors.New("order: invalid order")
var ErrNotFound = errors.New("order: not found")

type Repository interface {
	List(ctx context.Context) ([]Order, error)
	GetByID(ctx context.Context, id int64) (Order, error)
	Create(ctx context.Context, o Order) (Order, error)
}

type Service struct {
	Repository Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repository: repo}
}

func (s *Service) List(ctx context.Context) ([]Order, error) {
	return s.Repository.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int64) (Order, error) {
	return s.Repository.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, o Order) (Order, error) {
	if o.UserID == 0 || len(o.Items) == 0 {
		return Order{}, ErrInvalid
	}
	for _, item := range o.Items {
		if item.ListingID == 0 || item.Quantity <= 0 || item.Price < 0 {
			return Order{}, ErrInvalid
		}
	}
	return s.Repository.Create(ctx, o)
}
