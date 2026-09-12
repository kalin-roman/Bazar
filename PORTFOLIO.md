# Bazar

[![Go](https://github.com/kalin-roman/Bazar/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/kalin-roman/Bazar/actions/workflows/go.yml)

**Live**: [bazar-xxjl.onrender.com](https://bazar-xxjl.onrender.com) —
a real, deployed API (Render + Supabase Postgres, Supabase-issued JWT
auth). Every route requires a valid token, so a bare visit to the URL
correctly returns `401`, not a page — see below for what's actually
running behind it.

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
  domain packages (`product`, `category`, `order`, `user`) separate
  business logic (`Service`) from persistence (`Repository`, an
  interface owned by the consumer) — the service layer has no idea
  whether it's talking to a real database or a test double.
- **Test-driven habits**: table-driven unit tests against hand-written
  in-memory fake repositories, covering every validation branch, with
  no database dependency required to run the suite — for all four
  domains.
- **Schema design and real persistence**: hand-written SQL migrations
  (`golang-migrate`), including a deliberate normalization decision —
  storing product images in their own table rather than a denormalized
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
  library plumbing**: `internal/auth` verifies Supabase-issued ES256
  JWTs against keys fetched from Supabase's live JWKS endpoint (public-key,
  asymmetric verification — no shared secret), explicitly guarding
  against the algorithm-confusion attack class (rejecting a forged
  token that claims a different/no signing algorithm) rather than
  trusting the token's own header. Originally built and tested against
  HS256 on the (incorrect) assumption Supabase used shared-secret
  signing; when the real deployed token turned out to be ES256, the
  mismatch was diagnosed from first principles (decoding the token's
  header by hand, checking it against the JWKS endpoint) and the
  verifier was rewritten for real asymmetric verification — a genuine
  debugging story, not just a feature list item.
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
- **Actually deployed, not just running on localhost**: live on
  Render, backed by a real Supabase Postgres database, real
  Supabase-issued auth tokens verified against Supabase's live JWKS
  endpoint. Getting there surfaced and required fixing several real,
  disparate infrastructure problems in sequence — a hosting platform
  requiring payment info, a build system defaulting to the wrong
  runtime, an IPv4/IPv6 connection-pooling mismatch, the ES256 auth
  rewrite above, and a pre-existing, differently-shaped `users` table
  in the target database that had been silently halting every schema
  migration after it. Each was diagnosed from the actual error/log
  output, not guessed — see `docs/GO_LEARNING_LOG.md` for the full
  sequence.
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
│   ├── product/                # Product domain type + Repository/Service + tests
│   ├── category/                # Category domain type + Repository/Service + tests
│   ├── order/                    # Order domain type + Repository/Service + tests
│   ├── user/                      # User domain type + Repository/Service + tests
│   ├── auth/                       # Supabase JWT verification (ES256/JWKS)
│   ├── storage/postgres/            # Real pgx-backed Repository implementations, one file per domain
│   └── http/
│       ├── categoryhttp/               # Handler + router for /categories
│       ├── producthttp/                 # Handler + router for /products
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
- A plain struct with no framework annotations (`Product`, `Category`,
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
- Backend, all four domains (`product`, `category`, `order`, `user`):
  domain types, `Repository`/`Service` layers, full fake-backed test
  suites, real `pgx` `Repository` implementations validated against
  live Postgres, and HTTP handlers + routers.
- All SQL migrations (`categories`; `products` + `product_images`;
  `app_users`; `orders` + `order_items`), applied and rolled back
  cleanly against a real Postgres instance in both directions.
- JWT verification (`internal/auth`) against Supabase's real ES256/
  JWKS tokens, including the algorithm-confusion defense — verified
  live against Supabase's real JWKS endpoint and a real signed token
  (no committed unit tests for this package; correctness proven
  against the real external system instead throughout this project).
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
- **Deployed and live**: [bazar-xxjl.onrender.com](https://bazar-xxjl.onrender.com),
  on Render, backed by a real Supabase Postgres database. All four
  domains verified end-to-end against the live URL with a real,
  freshly-signed-in Supabase user's token — `200` on every route,
  `401` with no token.

**In progress:**
- `internal/platform/{logger,database,middleware}` — optional
  organizational polish (regrouping existing, already-working pieces),
  not functionally required.
- Authorization covers `order`/`user`; `category`/`product` have no
  owner concept to check (not user-scoped resources), so there's
  nothing analogous needed there.
- No committed test coverage for HTTP handlers or `internal/auth` —
  both verified live/via throwaway harnesses throughout instead of
  permanent `_test.go` files.

**Not started:**
- Reconnecting the frontend to the now-live backend instead of mock
  data.

In short: the backend's core roadmap is complete, live-proven, and
now actually deployed — domain logic, persistence, HTTP, real
Supabase auth verification, the middleware chain, authorization
checks on every user-owned resource, CI, and a real running
production URL are all in place. What's left is optional polish, a
bit more test coverage, and the separate, larger effort of
reconnecting the frontend.
