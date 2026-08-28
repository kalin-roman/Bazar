package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kalin-roman/Bazar/internal/listing"
)

type ListingRepository struct {
	ConnectionPool *pgxpool.Pool
}

var _ listing.Repository = (*ListingRepository)(nil)

func NewListingRepository(pool *pgxpool.Pool) *ListingRepository {
	return &ListingRepository{ConnectionPool: pool}
}

func (r *ListingRepository) List(ctx context.Context) ([]listing.Listing, error) {
	rows, err := r.ConnectionPool.Query(ctx,
		"select id, category_id, title, slug, price_cents, hero_image_url, max_quantity from listings")
	if err != nil {
		return nil, fmt.Errorf("list listings: %w", err)
	}
	defer rows.Close()

	var listings []listing.Listing
	for rows.Next() {
		var l listing.Listing
		if err := rows.Scan(&l.ID, &l.CategoryID, &l.Title, &l.Slug, &l.PriceCents, &l.HeroImageURL, &l.MaxQuantity); err != nil {
			return nil, fmt.Errorf("list listings: %w", err)
		}
		listings = append(listings, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list listings: %w", err)
	}

	return listings, nil
}

func (r *ListingRepository) GetBySlug(ctx context.Context, slug string) (listing.Listing, error) {
	row := r.ConnectionPool.QueryRow(ctx,
		"select id, category_id, title, slug, price_cents, hero_image_url, max_quantity from listings where slug = $1",
		slug)

	var l listing.Listing
	err := row.Scan(&l.ID, &l.CategoryID, &l.Title, &l.Slug, &l.PriceCents, &l.HeroImageURL, &l.MaxQuantity)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return listing.Listing{}, listing.ErrNotFound
		}
		return listing.Listing{}, fmt.Errorf("get listing by slug: %w", err)
	}

	return l, nil
}

func (r *ListingRepository) Create(ctx context.Context, l listing.Listing) (listing.Listing, error) {
	row := r.ConnectionPool.QueryRow(ctx,
		"insert into listings (category_id, title, slug, price_cents, hero_image_url, max_quantity) values ($1, $2, $3, $4, $5, $6) returning id",
		l.CategoryID, l.Title, l.Slug, l.PriceCents, l.HeroImageURL, l.MaxQuantity,
	)

	if err := row.Scan(&l.ID); err != nil {
		return listing.Listing{}, fmt.Errorf("create listing: %w", err)
	}

	return l, nil
}
