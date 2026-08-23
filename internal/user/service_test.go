package user

import (
	"context"
	"errors"
	"testing"
)

// fakeRepository is an in-memory Repository, used so Service can be
// tested without a real database.
type fakeRepository struct {
	users []User
}

var _ Repository = (*fakeRepository)(nil)

var errNotFound = errors.New("user: not found")

func (r *fakeRepository) List(ctx context.Context) ([]User, error) {
	return r.users, nil
}

func (r *fakeRepository) GetByID(ctx context.Context, id int64) (User, error) {
	for _, u := range r.users {
		if u.ID == id {
			return u, nil
		}
	}
	return User{}, errNotFound
}

func (r *fakeRepository) Create(ctx context.Context, u User) (User, error) {
	r.users = append(r.users, u)
	return u, nil
}

func TestServiceList(t *testing.T) {
	repo := &fakeRepository{users: []User{
		{ID: 1, FullName: "Ada Lovelace"},
		{ID: 2, FullName: "Grace Hopper"},
	}}
	s := NewService(repo)

	got, err := s.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("List returned %d users, want 2", len(got))
	}
}

func TestServiceGetByID(t *testing.T) {
	repo := &fakeRepository{users: []User{
		{ID: 1, FullName: "Ada Lovelace"},
	}}
	s := NewService(repo)

	tests := []struct {
		name         string
		id           int64
		wantFullName string
		wantErr      error
	}{
		{name: "found", id: 1, wantFullName: "Ada Lovelace"},
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
			if got.FullName != tc.wantFullName {
				t.Fatalf("got FullName %q, want %q", got.FullName, tc.wantFullName)
			}
		})
	}
}

func TestServiceCreate(t *testing.T) {
	valid := User{
		FullName: "Ada Lovelace",
		Email:    "ada@example.com",
	}

	tests := []struct {
		name    string
		mutate  func(u User) User
		wantErr error
	}{
		{
			name:   "valid",
			mutate: func(u User) User { return u },
		},
		{
			name:    "missing full name",
			mutate:  func(u User) User { u.FullName = ""; return u },
			wantErr: ErrInvalid,
		},
		{
			name:    "missing email",
			mutate:  func(u User) User { u.Email = ""; return u },
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
				if len(repo.users) != 0 {
					t.Fatalf("invalid Create should not have reached the repository")
				}
				return
			}
			if err != nil {
				t.Fatalf("Create returned error: %v", err)
			}
			if got.Email != valid.Email {
				t.Fatalf("got Email %q, want %q", got.Email, valid.Email)
			}
			if len(repo.users) != 1 {
				t.Fatalf("expected Create to store the user in the repository")
			}
		})
	}
}
