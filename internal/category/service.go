package category

import (
	"context"
	"errors"
)

var ErrInvalid = errors.New("category: invalid category")
var ErrNotFound = errors.New("category: not found")

type Repository interface {
	List(ctx context.Context) ([]Category, error)
	GetBySlug(ctx context.Context, slug string) (Category, error)
	Create(ctx context.Context, l Category) (Category, error)
}

type Service struct {
	Repository Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repository: repo}
}

func (s *Service) List(ctx context.Context) ([]Category, error) {
	return s.Repository.List(ctx)
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (Category, error) {
	return s.Repository.GetBySlug(ctx, slug)
}

func (s *Service) Create(ctx context.Context, l Category) (Category, error) {
	if len(l.ImageURL) == 0 || l.Slug == "" || l.Name == "" {
		return Category{}, ErrInvalid
	}
	return s.Repository.Create(ctx, l)
}
