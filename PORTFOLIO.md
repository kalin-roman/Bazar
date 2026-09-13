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
  system) — request logging, a closure-based auth-checking middleware
  (holds the JWKS verifier, propagates the verified user ID via
  `context.WithValue` using an unexported key type to avoid
  collisions), and CORS, wired in a specific, understood order rather
  than an arbitrary one: CORS outermost (a preflight request never
  carries a token, so it has to run before Auth even gets a chance to
  reject it), Auth next (a request with a missing/invalid token is
  rejected before Logging ever sees it), Logging innermost — verified
  live via the running server's own log output, not just by reading
  the code.
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
  rewrite above, a pre-existing, differently-shaped `users` table in
  the target database that had been silently halting every schema
  migration after it, and (found only once the frontend actually
  tried to call the API from a browser) a missing CORS layer — curl
  and a native mobile app's fetch never enforce CORS, so it was
  invisible until tested where it mattered. Each was diagnosed from
  the actual error/log output, not guessed — see
  `docs/GO_LEARNING_LOG.md` for the full sequence.
- **A real checkout endpoint that doesn't trust the client**:
  `POST /orders` accepts only a product ID and quantity per line item
  — never a price. The price actually charged is looked up
  server-side, at the moment of purchase, from the product's own
  current record. Verified with a dedicated test proving a
  maliciously low client-supplied price would be ignored even if one
  were sent, and confirmed live against the production database (the
  order row's stored price matches the product's real price, not
  anything the request could have claimed).
- **A fully reconnected React Native frontend**: real Supabase
  session auth (replacing an earlier local mock), a typed API client
  attaching the session's token to every request, and every screen —
  product/category browsing, cart, checkout, order history and detail
  — wired to the live API instead of static mock data. Verified by
  actually driving the app in a browser against the live backend:
  signing in, adding to cart, checking out, and confirming the
  resulting order independently in the database — not just reading
  the code. Two real, pre-existing bugs were found and fixed in the
  process: a Rules-of-Hooks violation (a hook called after an early
  conditional return — a real crash risk) and a session-storage
  library with no web implementation, silently broken until this was
  the first code path to actually exercise it.
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
│       ├── orderhttp/                    # Handler + router for /orders, incl. POST checkout
│       └── userhttp/                      # Handler + router for /users
├── internal/platform/            # Cross-cutting infrastructure, not domain logic
│   ├── database/                   # Postgres connection pool setup (pgx)
│   ├── middleware/                  # Hand-rolled net/http middleware (logging, JWT auth, CORS)
│   └── logger/                       # Centralized error logging
├── internal/config/                # Env-var config loading
├── migrations/                  # Hand-written SQL schema migrations (golang-migrate)
├── cmd/server/                    # Entry point — wired up, runs a real server
├── cmd/seed-temp/                  # Disposable script to seed real catalog data
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

**Done — the whole thing, backend and frontend, live in production:**
- All four backend domains (`product`, `category`, `order`, `user`):
  domain types, `Repository`/`Service` layers, fake-backed unit tests,
  real `pgx` `Repository` implementations, and HTTP handlers/routers
  — including `POST /orders` (checkout), which resolves each line
  item's price server-side rather than trusting the client.
- Handler-level tests (`httptest`, no real database needed) covering
  `category`'s and `order`'s HTTP layer, including the authorization
  matrix (owner vs. non-owner vs. nonexistent resource) and the
  checkout endpoint's price-authority guarantee.
- All SQL migrations (`categories`; `products` + `product_images`;
  `app_users`; `orders` + `order_items`, including `status` and
  `created_at`), applied and rolled back cleanly against both a local
  and the real production Postgres instance.
- JWT verification (`internal/auth`) against Supabase's real ES256/
  JWKS tokens, including the algorithm-confusion defense — verified
  live against Supabase's real JWKS endpoint and a real signed token.
- Three hand-rolled middlewares (request logging, JWT auth-checking,
  CORS), composed in a deliberate, understood order — CORS outermost
  of all (a preflight request never carries a token, so it has to run
  before Auth), Auth next, Logging innermost.
- `internal/platform/` — infrastructure code (`database`, `middleware`,
  `logger`) separated from domain logic, its own package tree.
- Authorization on `order.GetByID` and `user.GetByID` — ownership
  checks between the verified user and the requested resource, with a
  deliberate `404` (not `403`) for both "doesn't exist" and "exists
  but isn't yours," to avoid leaking which orders/users exist.
- CI (GitHub Actions: build/vet/gofmt/test on every push/PR), green.
- **Deployed and live**: [bazar-xxjl.onrender.com](https://bazar-xxjl.onrender.com),
  on Render, backed by a real Supabase Postgres database with real
  seeded catalog data.
- **The frontend is reconnected, end to end**: real Supabase session
  auth, a typed API client, and every screen — browsing, cart,
  checkout, order history/detail — wired to the live API. Proven by
  actually driving the deployed app in a browser: signing in with a
  real account, browsing real products, checking out, and confirming
  the resulting order independently in the production database.

**Known, deliberate gaps, not oversights:**
- No real image hosting — product photos are placeholder URLs
  (picsum.photos), seeded via `cmd/seed-temp`; building real upload
  storage was out of scope for this pass.
- `product`/`user` have no HTTP-level test coverage yet (`category`
  and `order` do) — `product` has no owner concept to test against
  anyway; `user`'s is the same shape as `order`'s already-tested one,
  just not mirrored into a test file yet.
- The backend's JSON responses currently use Go's exported field
  names as-is (`ID`, `PriceCents`, …) rather than conventional
  camelCase `json` tags — the frontend's types match this deliberately
  rather than silently diverging from what's actually deployed.

In short: this isn't a partial slice — domain logic, persistence,
HTTP, real Supabase auth (both verifying incoming tokens and issuing
them via the reconnected frontend), authorization, a checkout flow
that doesn't trust the client, CI, and a live production deployment
serving a real, working mobile-web app are all in place and verified
against the real running system, not just written and assumed.
