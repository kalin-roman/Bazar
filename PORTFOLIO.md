# Bazar

[![Go](https://github.com/kalin-roman/Bazar/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/kalin-roman/Bazar/actions/workflows/go.yml)

A full-stack mobile marketplace project: a React Native / Expo frontend
paired with a Go backend built from scratch, by hand, as a structured
learning project in backend engineering and Go.

This file is a technical overview for anyone evaluating the codebase
(e.g. as part of a job application). The user-facing `README.md`
covers frontend setup/run instructions; this one covers what the
project actually demonstrates and its current state.

## What this project demonstrates

- **Go fundamentals done deliberately, not copy-pasted**: package
  structure (`internal/`), implicit/structural interfaces, idiomatic
  error handling (sentinel errors checked via `errors.Is`, wrapped with
  `%w`), and Go's naming/casing conventions.
- **The Repository/Service pattern, applied consistently**: all four
  domain packages (`listing`, `category`, `order`, `user`) separate
  business logic (`Service`) from persistence (`Repository`, an
  interface owned by the consumer) — the service layer has no idea
  whether it's talking to a real database or a test double.
- **Test-driven habits**: table-driven unit tests against hand-written
  in-memory fake repositories, covering every validation branch, with
  no database dependency required to run the suite — for all four
  domains.
- **Schema design and real persistence**: hand-written SQL migrations
  (`golang-migrate`), including a deliberate normalization decision —
  storing listing images in their own table rather than a denormalized
  array column, specifically to support per-image editing/ordering
  later — plus real `pgx`-backed `Repository` implementations for all
  four domains, each validated end-to-end against a live Postgres
  instance (not just compiled — actually run, with rows written and
  read back). `order`'s implementation uses a real transaction
  (`pool.Begin`/`Commit`/deferred `Rollback`) so a multi-table write
  (an order plus its line items) can't land partially committed.
- **HTTP layer on the standard library, no framework**: Go 1.22+
  pattern-based `net/http` routing (`GET /categories`, etc.) with a
  clean handler → service → repository layering — handlers never touch
  the database directly. Implemented for all four domains.
- **JWT verification with an actual security rationale, not just
  library plumbing**: `internal/auth` verifies Supabase-issued HS256
  JWTs and explicitly guards against the algorithm-confusion attack
  class (rejecting a forged token that claims a different/no signing
  algorithm) rather than trusting the token's own header — verified
  with tests covering a valid token, an expired token, a wrong-secret
  token, and a forged `alg:none` token.
- **Hand-rolled middleware, composed deliberately**: the
  `net/http.Handler` wrap-and-delegate pattern (no framework middleware
  system) — a request-logging middleware and a closure-based
  auth-checking middleware (carries a secret, propagates the verified
  user ID via `context.WithValue` using an unexported key type to avoid
  collisions), wired together in a specific, understood order: auth
  runs outermost, so a request with a missing/invalid token is rejected
  before the logging middleware ever sees it — verified live via the
  running server's own log output, not just by reading the code.
- **A fully wired, live-tested server**: `cmd/server/main.go` composes
  every layer above — config, DB pool, all four domains'
  repository/service/handler stacks, one shared `http.ServeMux`, and
  the middleware chain — into one running `http.Server`. Proven with a
  real end-to-end smoke test against live Postgres: all four domains
  correctly return `401` for missing/invalid tokens and `200` for a
  valid one.
- **Authorization, not just authentication**: both `order.GetByID` and
  `user.GetByID` check that the verified requester (from the JWT)
  actually owns the resource being fetched, not just that they're
  logged in. A `404` (not `403`) is returned for both "doesn't exist"
  and "exists but isn't yours," deliberately — avoiding resource
  enumeration (confirming an ID exists to someone who shouldn't see
  it), the same reasoning GitHub uses for private repos. Verified live
  with two distinct users' JWTs against a real running server, for
  both domains: the owner gets `200`, a different authenticated user
  gets the identical `404` a nonexistent resource would return.
- **CI**: a GitHub Actions workflow runs `go build`, `go vet`, a
  `gofmt` check, and `go test` on every push/PR — verified green
  against a real run, not just written and assumed.
- **A modern React Native app**: Expo Router v6 file-based routing,
  TypeScript throughout, Zustand for client state, Supabase for auth
  and persistence (frontend side).
- **A documented learning process**: [`docs/GO_LEARNING_LOG.md`](docs/GO_LEARNING_LOG.md)
  is a running, lesson-by-lesson log of the Go backend build — what
  was taught, what was built, what bugs came up in review and why. It
  was kept up throughout rather than written retroactively, so it's an
  honest record of the actual learning process, not a highlight reel.

## Architecture

```
Bazar/
├── web/                       # React Native / Expo frontend (see README.md)
├── internal/
│   ├── listing/                # Listing domain type + Repository/Service + tests
│   ├── category/                # Category domain type + Repository/Service + tests
│   ├── order/                    # Order domain type + Repository/Service + tests
│   ├── user/                      # User domain type + Repository/Service + tests
│   ├── auth/                       # Supabase JWT verification (HS256)
│   ├── storage/postgres/            # Real pgx-backed Repository implementations, one file per domain
│   └── http/
│       ├── categoryhttp/               # Handler + router for /categories
│       ├── listinghttp/                 # Handler + router for /listings
│       ├── orderhttp/                    # Handler + router for /orders
│       ├── userhttp/                      # Handler + router for /users
│       └── middleware/                     # Hand-rolled net/http middleware (logging, JWT auth)
├── migrations/                  # Hand-written SQL schema migrations (golang-migrate)
├── db/                           # Postgres connection pool setup (pgx)
├── internal/config/                # Env-var config loading
├── cmd/server/                    # Entry point — wired up, runs a real server
└── docs/GO_LEARNING_LOG.md          # Lesson-by-lesson build log
```

Each backend domain package follows the same shape:
- A plain struct with no framework annotations (`Listing`, `Category`,
  `Order`, `User`).
- A `Repository` interface defining what persistence operations are
  needed — owned by the consumer (the domain package), not by whatever
  eventually implements it.
- A `Service` holding business logic (validation, etc.) that depends
  only on the `Repository` interface, never a concrete database type.
- A hand-written in-memory fake implementing `Repository`, used in a
  table-driven test suite with no database required.
- A real `pgx`-backed implementation of that same `Repository`
  interface in `internal/storage/postgres`, validated against a live
  database.
- An HTTP handler and router in `internal/http/<domain>http` exposing
  it over `net/http`.

## Current status

**Done:**
- Frontend: browsable end-to-end against local mock data (products,
  categories, orders). Supabase auth code exists and is fully
  functional but currently disconnected by design, so every screen is
  reachable without signing in during frontend development.
- Backend, all four domains (`listing`, `category`, `order`, `user`):
  domain types, `Repository`/`Service` layers, full fake-backed test
  suites, real `pgx` `Repository` implementations validated against
  live Postgres, and HTTP handlers + routers.
- All four SQL migrations (`categories`; `listings` + `listing_images`;
  `users`; `orders` + `order_items`), applied and rolled back cleanly
  against a real Postgres instance in both directions.
- JWT verification (`internal/auth`), including the algorithm-confusion
  defense, covered by tests.
- Both hand-rolled middlewares (request logging, JWT auth-checking),
  composed together in a deliberate order.
- Env-var-based config loading (`internal/config`).
- All four domains' routers composed onto one shared `http.ServeMux`.
- `cmd/server/main.go` — fully wired: config → DB pool → all four
  domains' repository/service/handler stacks → shared mux → middleware
  chain → running `http.Server`. Live end-to-end proof: every domain
  correctly returns `401` for a missing/invalid token and `200` for a
  valid one, confirmed against a real running Postgres instance.

- Authorization on `order.GetByID` and `user.GetByID` — ownership
  checks between the verified user and the requested resource,
  live-tested with two distinct users against a real running server,
  for both domains.
- CI (GitHub Actions: build/vet/gofmt/test on every push/PR), green
  and running.

**In progress:**
- `internal/platform/{logger,database,middleware}` — optional
  organizational polish (regrouping existing, already-working pieces),
  not functionally required.
- Authorization covers `order`/`user`; `category`/`listing` have no
  owner concept to check (not user-scoped resources), so there's
  nothing analogous needed there.

**Not started:**
- Reconnecting the frontend to a live backend instead of mock data.
- Actually deploying the backend somewhere reachable.

In short: the backend's core roadmap is complete and live-proven —
domain logic, persistence, HTTP, auth verification, the middleware
chain, real authorization checks on every user-owned resource, and CI
are all in place. What's left is optional polish, deployment, and the
separate, larger effort of reconnecting the frontend.
