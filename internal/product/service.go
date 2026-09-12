package product

import (
	"context"
	"errors"
)

var ErrInvalid = errors.New("product: invalid product")
var ErrNotFound = errors.New("product: not found")

type Repository interface {
	List(ctx context.Context) ([]Product, error)
	GetBySlug(ctx context.Context, slug string) (Product, error)
	GetByID(ctx context.Context, id int64) (Product, error)
	Create(ctx context.Context, p Product) (Product, error)
}

type Service struct {
	Repository Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repository: repo}
}

func (s *Service) List(ctx context.Context) ([]Product, error) {
	return s.Repository.List(ctx)
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (Product, error) {
	return s.Repository.GetBySlug(ctx, slug)
}

func (s *Service) GetByID(ctx context.Context, id int64) (Product, error) {
	return s.Repository.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, p Product) (Product, error) {
	if p.CategoryID == 0 || p.MaxQuantity < 0 || p.PriceCents < 0 {
		return Product{}, ErrInvalid
	}
	if p.HeroImageURL == "" || len(p.ImagesURL) == 0 || p.Slug == "" || p.Title == "" {
		return Product{}, ErrInvalid
	}
	return s.Repository.Create(ctx, p)
}
