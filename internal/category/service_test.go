package category

import (
	"context"
	"errors"
	"testing"
)

// fakeRepository is an in-memory Repository, used so Service can be
// tested without a real database.
type fakeRepository struct {
	categories []Category
}

var _ Repository = (*fakeRepository)(nil)

var errNotFound = errors.New("category: not found")

func (r *fakeRepository) List(ctx context.Context) ([]Category, error) {
	return r.categories, nil
}

func (r *fakeRepository) GetBySlug(ctx context.Context, slug string) (Category, error) {
	for _, c := range r.categories {
		if c.Slug == slug {
			return c, nil
		}
	}
	return Category{}, errNotFound
}

func (r *fakeRepository) Create(ctx context.Context, c Category) (Category, error) {
	r.categories = append(r.categories, c)
	return c, nil
}

func TestServiceList(t *testing.T) {
	repo := &fakeRepository{categories: []Category{
		{ID: 1, Slug: "furniture"},
		{ID: 2, Slug: "lighting"},
	}}
	s := NewService(repo)

	got, err := s.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("List returned %d categories, want 2", len(got))
	}
}

func TestServiceGetBySlug(t *testing.T) {
	repo := &fakeRepository{categories: []Category{
		{ID: 1, Slug: "furniture", Name: "Furniture"},
	}}
	s := NewService(repo)

	tests := []struct {
		name    string
		slug    string
		wantID  int64
		wantErr error
	}{
		{name: "found", slug: "furniture", wantID: 1},
		{name: "missing", slug: "lighting", wantErr: errNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := s.GetBySlug(context.Background(), tc.slug)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("got err %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetBySlug returned error: %v", err)
			}
			if got.ID != tc.wantID {
				t.Fatalf("got ID %d, want %d", got.ID, tc.wantID)
			}
		})
	}
}

func TestServiceCreate(t *testing.T) {
	valid := Category{
		Name:     "Furniture",
		Slug:     "furniture",
		ImageURL: "furniture.png",
	}

	tests := []struct {
		name    string
		mutate  func(c Category) Category
		wantErr error
	}{
		{
			name:   "valid",
			mutate: func(c Category) Category { return c },
		},
		{
			name:    "missing image",
			mutate:  func(c Category) Category { c.ImageURL = ""; return c },
			wantErr: ErrInvalid,
		},
		{
			name:    "missing slug",
			mutate:  func(c Category) Category { c.Slug = ""; return c },
			wantErr: ErrInvalid,
		},
		{
			name:    "missing name",
			mutate:  func(c Category) Category { c.Name = ""; return c },
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
				if len(repo.categories) != 0 {
					t.Fatalf("invalid Create should not have reached the repository")
				}
				return
			}
			if err != nil {
				t.Fatalf("Create returned error: %v", err)
			}
			if got.Slug != valid.Slug {
				t.Fatalf("got slug %q, want %q", got.Slug, valid.Slug)
			}
			if len(repo.categories) != 1 {
				t.Fatalf("expected Create to store the category in the repository")
			}
		})
	}
}
