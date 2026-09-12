package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kalin-roman/Bazar/internal/product"
)

type ProductRepository struct {
	ConnectionPool *pgxpool.Pool
}

var _ product.Repository = (*ProductRepository)(nil)

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{ConnectionPool: pool}
}

func (r *ProductRepository) List(ctx context.Context) ([]product.Product, error) {
	rows, err := r.ConnectionPool.Query(ctx,
		"select id, category_id, title, slug, price_cents, hero_image_url, max_quantity from products")
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	var products []product.Product
	for rows.Next() {
		var p product.Product
		if err := rows.Scan(&p.ID, &p.CategoryID, &p.Title, &p.Slug, &p.PriceCents, &p.HeroImageURL, &p.MaxQuantity); err != nil {
			return nil, fmt.Errorf("list products: %w", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}

	return products, nil
}

func (r *ProductRepository) GetBySlug(ctx context.Context, slug string) (product.Product, error) {
	row := r.ConnectionPool.QueryRow(ctx,
		"select id, category_id, title, slug, price_cents, hero_image_url, max_quantity from products where slug = $1",
		slug)

	var p product.Product
	err := row.Scan(&p.ID, &p.CategoryID, &p.Title, &p.Slug, &p.PriceCents, &p.HeroImageURL, &p.MaxQuantity)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return product.Product{}, product.ErrNotFound
		}
		return product.Product{}, fmt.Errorf("get product by slug: %w", err)
	}

	return p, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id int64) (product.Product, error) {
	row := r.ConnectionPool.QueryRow(ctx,
		"select id, category_id, title, slug, price_cents, hero_image_url, max_quantity from products where id = $1",
		id)

	var p product.Product
	err := row.Scan(&p.ID, &p.CategoryID, &p.Title, &p.Slug, &p.PriceCents, &p.HeroImageURL, &p.MaxQuantity)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return product.Product{}, product.ErrNotFound
		}
		return product.Product{}, fmt.Errorf("get product by id: %w", err)
	}

	return p, nil
}

func (r *ProductRepository) Create(ctx context.Context, p product.Product) (product.Product, error) {
	row := r.ConnectionPool.QueryRow(ctx,
		"insert into products (category_id, title, slug, price_cents, hero_image_url, max_quantity) values ($1, $2, $3, $4, $5, $6) returning id",
		p.CategoryID, p.Title, p.Slug, p.PriceCents, p.HeroImageURL, p.MaxQuantity,
	)

	if err := row.Scan(&p.ID); err != nil {
		return product.Product{}, fmt.Errorf("create product: %w", err)
	}

	return p, nil
}
