package user

import (
	"context"
	"errors"
)

var ErrInvalid = errors.New("user: invalid user")

type Repository interface {
	List(ctx context.Context) ([]User, error)
	GetByID(ctx context.Context, id int64) (User, error)
	Create(ctx context.Context, o User) (User, error)
}

type Service struct {
	Repository Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repository: repo}
}

func (s *Service) List(ctx context.Context) ([]User, error) {
	return s.Repository.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int64) (User, error) {
	return s.Repository.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, o User) (User, error) {
	if len(o.FullName) == 0 || len(o.Email) == 0 {
		return User{}, ErrInvalid
	}

	return s.Repository.Create(ctx, o)
}
