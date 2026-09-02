# Bazar

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
- **Hand-rolled middleware**: the `net/http.Handler`
  wrap-and-delegate pattern (no framework middleware system), currently
  demonstrated with a request-logging middleware; an auth-checking
  middleware is designed (closure-based, to carry a secret; propagates
  the verified user ID via `context.WithValue` using an unexported key
  type to avoid collisions) and is the next piece to implement.
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
│       └── middleware/                     # Hand-rolled net/http middleware (logging; auth designed, not yet written)
├── migrations/                  # Hand-written SQL schema migrations (golang-migrate)
├── db/                           # Postgres connection pool setup (pgx)
├── cmd/server/                    # Entry point — not wired up yet
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
- First hand-rolled middleware (request logging), proving the pattern.

**In progress:**
- Auth middleware — design finalized, not yet implemented.
- Composing the four domains' independent routers into a single
  `http.Handler` a real `http.Server` can serve (currently each domain
  builds its own separate `*http.ServeMux`; these don't yet merge into
  one).
- `internal/config` and `internal/platform/{logger,database,middleware}`.
- `cmd/server/main.go` — wiring everything into an actual running
  server. This is the last major step; once it's in, the backend goes
  from "a set of proven, independently-tested pieces" to "a server you
  can actually run and hit with curl."

**Not started:**
- Reconnecting the frontend to a live backend instead of mock data.

In short: this is an active work-in-progress. Every layer of the
backend — domain logic, persistence, HTTP, auth verification, and the
middleware pattern — has been built and independently proven (tests,
or live validation against a real Postgres instance), but the final
assembly step that turns those proven pieces into one running server
hasn't happened yet.
