package main

import (
	"context"
	"log"
	"net/http"

	"github.com/kalin-roman/Bazar/db"
	"github.com/kalin-roman/Bazar/internal/auth"
	"github.com/kalin-roman/Bazar/internal/category"
	"github.com/kalin-roman/Bazar/internal/config"
	"github.com/kalin-roman/Bazar/internal/http/categoryhttp"
	"github.com/kalin-roman/Bazar/internal/http/middleware"
	"github.com/kalin-roman/Bazar/internal/http/orderhttp"
	"github.com/kalin-roman/Bazar/internal/http/producthttp"
	"github.com/kalin-roman/Bazar/internal/http/userhttp"
	"github.com/kalin-roman/Bazar/internal/order"
	"github.com/kalin-roman/Bazar/internal/product"
	"github.com/kalin-roman/Bazar/internal/storage/postgres"
	"github.com/kalin-roman/Bazar/internal/user"
)

func main() {
	// A root context with no deadline/cancellation. Every pgx call below
	// takes a context so it can be cancelled or given a timeout later —
	// for now this just satisfies that requirement.
	ctx := context.Background()

	cfg, err := config.Load() // to get all of the connection links, codes, and JWT secretes

	if err != nil {
		log.Fatal(err) // config.Load already failed, nothing downstream can work without it — stop immediately
	}

	pool, err := db.New(ctx, cfg.ConnectionString) // to opecn connection to the connection pool with the certaing key

	if err != nil {
		log.Fatal(err) // can't serve anything without a working DB pool — stop immediately, same reasoning as above
	}

	// Fetches Supabase's public signing key(s) once, at startup — same
	// fail-fast reasoning as everything above: nothing can verify a
	// single request without this, so a fetch failure here should stop
	// the server rather than let every request fail individually later.
	verifier, err := auth.NewVerifier(cfg.JWKSURL)
	if err != nil {
		log.Fatal(err)
	}
	// Runs when main() returns (any path), so the pool always gets
	// released cleanly instead of leaking connections.
	defer pool.Close()

	// categorie part

	// Talks to Postgres directly via pool. Satisfies category.Repository
	// (see the var _ category.Repository assertion in categories.go) —
	// nothing above this line knows or cares that it's pgx underneath.
	categoryRepo := postgres.NewCategorieRepository(pool)
	// Only knows the Repository interface, not pgx/pool at all. Holds
	// the validation + business logic (e.g. Create's field checks).
	categorySvc := category.NewService(categoryRepo)
	// Only knows *category.Service. Translates HTTP requests into
	// service calls and writes JSON responses — never touches the DB
	// or the Repository directly.
	categoryHandler := categoryhttp.NewCategoriesService(categorySvc)

	// product part

	// Talks to Postgres directly via pool. Satisfies product.Repository —
	// nothing above this line knows or cares that it's pgx underneath.
	productRepo := postgres.NewProductRepository(pool)
	// Only knows the Repository interface, not pgx/pool at all. Holds
	// the validation + business logic (e.g. Create's field checks).
	productSvc := product.NewService(productRepo)
	// Only knows *product.Service. Translates HTTP requests into
	// service calls and writes JSON responses — never touches the DB
	// or the Repository directly.
	productHandler := producthttp.NewProductService(productSvc)

	// Order part

	// Talks to Postgres directly via pool. Satisfies order.Repository —
	// nothing above this line knows or cares that it's pgx underneath.
	orderRepo := postgres.NewOrderRepository(pool)
	// Only knows the Repository interface, not pgx/pool at all. Holds
	// the validation + business logic (e.g. Create's field checks).
	orderSvc := order.NewService(orderRepo)
	// Only knows *order.Service. Translates HTTP requests into
	// service calls and writes JSON responses — never touches the DB
	// or the Repository directly.
	orderHandler := orderhttp.NewOrderService(orderSvc)

	// User Part

	// Talks to Postgres directly via pool. Satisfies user.Repository —
	// nothing above this line knows or cares that it's pgx underneath.
	userRepo := postgres.NewUserRepository(pool)
	// Only knows the Repository interface, not pgx/pool at all. Holds
	// the validation + business logic (e.g. Create's field checks).
	userSrv := user.NewService(userRepo)
	// Only knows *user.Service. Translates HTTP requests into
	// service calls and writes JSON responses — never touches the DB
	// or the Repository directly.
	userHandler := userhttp.NewUserService(userSrv)

	// One shared mux for the whole server. An http.Server needs exactly
	// one handler — this is what lets four independently-built domain
	// stacks end up served by a single running server instead of four.
	mux := http.NewServeMux()

	// Each call below mutates mux in place (RegisterRouter has no
	// return value) by adding that domain's route pattern(s) to it.
	//   - mux:            the shared *http.ServeMux every domain registers onto,
	//                      so all routes end up reachable from one handler.
	//   - <domain>Handler: that domain's *HandlesService — the thing whose
	//                      methods (e.g. List) actually get called once a
	//                      request matches the registered pattern.
	categoryhttp.RegisterRouter(mux, categoryHandler)
	producthttp.RegisterRouter(mux, productHandler)
	orderhttp.RegisterRouter(mux, orderHandler)
	userhttp.RegisterRouter(mux, userHandler)

	// handler starts as mux, then gets wrapped one layer at a time.
	// Whichever wrap happens LAST ends up OUTERMOST, meaning its code is
	// what actually runs first when a real request arrives — the source
	// order below is nesting order, not execution order.
	var handler http.Handler = mux
	// Wrapped first, so it ends up as the innermost layer: only reached
	// if Auth (wrapped next) calls next.ServeHTTP.
	handler = middleware.Logging(handler)
	// Wrapped last, so it's outermost: runs first on every request. A
	// request with a missing/invalid token gets rejected with 401 here,
	// before Logging (or mux) ever sees it — so only requests that pass
	// Auth get timed/logged. Deliberate trade-off, see the log for why.
	handler = middleware.Auth(verifier)(handler)

	// Blocks here for as long as the server runs. Only returns once it
	// stops — either from a real failure (e.g. port already in use) or
	// a graceful shutdown, neither of which exist yet, so any return
	// here means something went wrong.
	log.Fatal(http.ListenAndServe(":8080", handler))
}
