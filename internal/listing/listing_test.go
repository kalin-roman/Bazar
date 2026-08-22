package listing

import (
	"context"
	"errors"
	"testing"
)

// fakeRepository is an in-memory Repository, used so Service can be
// tested without a real database.
type fakeRepository struct {
	listings []Listing
}

var _ Repository = (*fakeRepository)(nil)

var errNotFound = errors.New("listing: not found")

func (r *fakeRepository) List(ctx context.Context) ([]Listing, error) {
	return r.listings, nil
}

func (r *fakeRepository) GetBySlug(ctx context.Context, slug string) (Listing, error) {
	for _, l := range r.listings {
		if l.Slug == slug {
			return l, nil
		}
	}
	return Listing{}, errNotFound
}

func (r *fakeRepository) Create(ctx context.Context, l Listing) (Listing, error) {
	r.listings = append(r.listings, l)
	return l, nil
}

func TestServiceList(t *testing.T) {
	repo := &fakeRepository{listings: []Listing{
		{ID: 1, Slug: "chair"},
		{ID: 2, Slug: "table"},
	}}
	s := NewService(repo)

	got, err := s.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("List returned %d listings, want 2", len(got))
	}
}

func TestServiceGetBySlug(t *testing.T) {
	repo := &fakeRepository{listings: []Listing{
		{ID: 1, Slug: "chair", Title: "Chair"},
	}}
	s := NewService(repo)

	tests := []struct {
		name    string
		slug    string
		wantID  int64
		wantErr error
	}{
		{name: "found", slug: "chair", wantID: 1},
		{name: "missing", slug: "sofa", wantErr: errNotFound},
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
	valid := Listing{
		CategoryID:   1,
		Title:        "Chair",
		Slug:         "chair",
		ImagesURL:    []string{"chair.png"},
		PriceCents:   1000,
		HeroImageURL: "chair-hero.png",
		MaxQuantity:  5,
	}

	tests := []struct {
		name    string
		mutate  func(l Listing) Listing
		wantErr error
	}{
		{
			name:   "valid",
			mutate: func(l Listing) Listing { return l },
		},
		{
			name:    "zero category",
			mutate:  func(l Listing) Listing { l.CategoryID = 0; return l },
			wantErr: ErrInvalid,
		},
		{
			name:    "negative max quantity",
			mutate:  func(l Listing) Listing { l.MaxQuantity = -1; return l },
			wantErr: ErrInvalid,
		},
		{
			name:    "negative price",
			mutate:  func(l Listing) Listing { l.PriceCents = -1; return l },
			wantErr: ErrInvalid,
		},
		{
			name:    "missing hero image",
			mutate:  func(l Listing) Listing { l.HeroImageURL = ""; return l },
			wantErr: ErrInvalid,
		},
		{
			name:    "no images",
			mutate:  func(l Listing) Listing { l.ImagesURL = nil; return l },
			wantErr: ErrInvalid,
		},
		{
			name:    "missing slug",
			mutate:  func(l Listing) Listing { l.Slug = ""; return l },
			wantErr: ErrInvalid,
		},
		{
			name:    "missing title",
			mutate:  func(l Listing) Listing { l.Title = ""; return l },
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
				if len(repo.listings) != 0 {
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
			if len(repo.listings) != 1 {
				t.Fatalf("expected Create to store the listing in the repository")
			}
		})
	}
}
