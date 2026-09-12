package categoryhttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kalin-roman/Bazar/internal/category"
)

// fakeRepository is a minimal in-memory category.Repository, just
// enough to drive HandlesService.List directly, without a real
// database or the mux/middleware in front of it.
type fakeRepository struct {
	categories []category.Category
	err        error
}

var _ category.Repository = (*fakeRepository)(nil)

func (r *fakeRepository) List(ctx context.Context) ([]category.Category, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.categories, nil
}

func (r *fakeRepository) GetBySlug(ctx context.Context, slug string) (category.Category, error) {
	return category.Category{}, errors.New("not used by this test")
}

func (r *fakeRepository) Create(ctx context.Context, c category.Category) (category.Category, error) {
	return category.Category{}, errors.New("not used by this test")
}

func TestList(t *testing.T) {
	tests := []struct {
		name       string
		repo       *fakeRepository
		wantStatus int
	}{
		{
			name: "success",
			repo: &fakeRepository{categories: []category.Category{
				{ID: 1, Name: "Chairs", Slug: "chairs", ImageURL: "chairs.png"},
				{ID: 2, Name: "Tables", Slug: "tables", ImageURL: "tables.png"},
			}},
			wantStatus: http.StatusOK,
		},
		{
			name:       "repository error",
			repo:       &fakeRepository{err: errors.New("db down")},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := category.NewService(tc.repo)
			h := NewCategoriesService(svc)

			r := httptest.NewRequest(http.MethodGet, "/categories", nil)
			w := httptest.NewRecorder()

			h.List(w, r)

			if w.Code != tc.wantStatus {
				t.Fatalf("got status %d, want %d", w.Code, tc.wantStatus)
			}

			if tc.wantStatus != http.StatusOK {
				return
			}

			var got []category.Category
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("decode response body: %v", err)
			}
			if len(got) != len(tc.repo.categories) {
				t.Fatalf("got %d categories, want %d", len(got), len(tc.repo.categories))
			}
		})
	}
}
