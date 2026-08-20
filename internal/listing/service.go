package listing

import (
	"context"
	"errors"
)

var ErrInvalid = errors.New("listing: invalid listing")

type Repository interface {
	List(ctx context.Context) ([]Listing, error)
	GetBySlug(ctx context.Context, slug string) (Listing, error)
	Create(ctx context.Context, l Listing) (Listing, error)
}

type Service struct {
	Repository Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repository: repo}
}

func (s *Service) List(ctx context.Context) ([]Listing, error) {
	return s.Repository.List(ctx)
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (Listing, error) {
	return s.Repository.GetBySlug(ctx, slug)
}

func (s *Service) Create(ctx context.Context, l Listing) (Listing, error) {
	if l.CategoryID == 0 || l.MaxQuantity < 0 || l.PriceCents < 0 {
		return Listing{}, ErrInvalid
	}
	if l.HeroImageURL == "" || len(l.ImagesURL) == 0 || l.Slug == "" || l.Title == "" {
		return Listing{}, ErrInvalid
	}
	return s.Repository.Create(ctx, l)
}
