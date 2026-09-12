package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kalin-roman/Bazar/internal/category"
	"github.com/kalin-roman/Bazar/internal/product"
	"github.com/kalin-roman/Bazar/internal/storage/postgres"
)

// Disposable seeding script — run once against a real database
// (local or Supabase) to populate a handful of real categories and
// products for the frontend to browse. Uses picsum.photos for
// placeholder images, since there's no real image hosting yet.
func main() {
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Fatal("set DATABASE_URL first")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	catRepo := postgres.NewCategorieRepository(pool)
	prodRepo := postgres.NewProductRepository(pool)

	type seedProduct struct {
		title       string
		slug        string
		priceCents  int64
		maxQuantity int64
	}

	categories := []struct {
		name     string
		slug     string
		products []seedProduct
	}{
		{
			name: "Furniture", slug: "furniture",
			products: []seedProduct{
				{"Oak Dining Chair", "oak-dining-chair", 8999, 10},
				{"Velvet Sofa", "velvet-sofa", 89999, 3},
			},
		},
		{
			name: "Electronics", slug: "electronics",
			products: []seedProduct{
				{"Wireless Headphones", "wireless-headphones", 12999, 20},
				{"Smart Watch", "smart-watch", 19999, 15},
			},
		},
		{
			name: "Clothing", slug: "clothing",
			products: []seedProduct{
				{"Denim Jacket", "denim-jacket", 6999, 25},
				{"Wool Scarf", "wool-scarf", 2499, 30},
			},
		},
	}

	for _, c := range categories {
		cat, err := catRepo.Create(ctx, category.Category{
			Name:     c.name,
			Slug:     c.slug,
			ImageURL: fmt.Sprintf("https://picsum.photos/seed/%s/600/400", c.slug),
		})
		if err != nil {
			log.Fatalf("create category %s: %v", c.slug, err)
		}
		fmt.Printf("category %-12s id=%d\n", c.slug, cat.ID)

		for _, p := range c.products {
			hero := fmt.Sprintf("https://picsum.photos/seed/%s/800/800", p.slug)
			prod, err := prodRepo.Create(ctx, product.Product{
				CategoryID:   cat.ID,
				Title:        p.title,
				Slug:         p.slug,
				PriceCents:   p.priceCents,
				HeroImageURL: hero,
				ImagesURL:    []string{hero},
				MaxQuantity:  p.maxQuantity,
			})
			if err != nil {
				log.Fatalf("create product %s: %v", p.slug, err)
			}
			fmt.Printf("  product %-24s id=%d price=$%.2f\n", p.slug, prod.ID, float64(p.priceCents)/100)
		}
	}

	fmt.Println("done")
}
