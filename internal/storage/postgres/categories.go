package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kalin-roman/Bazar/internal/category"
)

type Repository struct {
	ConnectionPool *pgxpool.Pool
}

var _ category.Repository = (*Repository)(nil)

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{ConnectionPool: pool}
}

func (r *Repository) GetBySlug(ctx context.Context, slug string) (category.Category, error) {
	row := r.ConnectionPool.QueryRow(ctx, "select id, name, slug, image_url from categories where slug = $1", slug)

	var c category.Category
	// Scan reads the row's columns, in order, into the given
	// destination pointers.
	err := row.Scan(&c.ID, &c.Name, &c.Slug, &c.ImageURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return category.Category{}, category.ErrNotFound
		}
		return category.Category{}, fmt.Errorf("get category by slug: %w", err)
	}

	return c, nil
}

func (r *Repository) List(ctx context.Context) ([]category.Category, error) {
	rows, err := r.ConnectionPool.Query(ctx, "select id, name, slug, image_url from categories")
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var categories []category.Category
	for rows.Next() {
		var c category.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.ImageURL); err != nil {
			return nil, fmt.Errorf("list categories: %w", err)
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}

	return categories, nil
}

func (r *Repository) Create(ctx context.Context, c category.Category) (category.Category, error) {
	row := r.ConnectionPool.QueryRow(ctx,
		"insert into categories (name, slug, image_url) values ($1, $2, $3) returning id",
		c.Name, c.Slug, c.ImageURL,
	)

	if err := row.Scan(&c.ID); err != nil {
		return category.Category{}, fmt.Errorf("create category: %w", err)
	}

	return c, nil
}
