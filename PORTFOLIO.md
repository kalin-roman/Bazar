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
  error handling (sentinel errors checked via `errors.Is`), and Go's
  naming/casing conventions.
- **The Repository/Service pattern**: each domain package (`listing`,
  `category`, `order`) separates business logic (`Service`) from
  persistence (`Repository`, an interface) — the service layer has no
  idea whether it's talking to a real database or a test double.
- **Test-driven habits**: table-driven unit tests against hand-written
  in-memory fake repositories, covering every validation branch, with
  no database dependency required to run the suite.
- **Schema design**: hand-written SQL migrations (via
  `golang-migrate`), including a deliberate normalization decision —
  storing listing images in their own table rather than a denormalized
  array column, specifically to support per-image editing/ordering
  later.
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
├── web/                # React Native / Expo frontend (see README.md)
├── internal/            # Go backend — domain logic, framework-free
│   ├── listing/         # Listing domain type + Repository/Service + tests
│   ├── category/        # Category domain type + Repository/Service + tests
│   └── order/            # Order domain type + Repository/Service + tests
├── migrations/           # Hand-written SQL schema migrations (golang-migrate)
├── db/                    # Postgres connection (pgx) — in progress
├── cmd/server/            # Entry point — not wired up yet
└── docs/GO_LEARNING_LOG.md  # Lesson-by-lesson build log
```

Each backend domain package follows the same shape:
- A plain struct with no framework annotations (`Listing`, `Category`,
  `Order`).
- A `Repository` interface defining what persistence operations are
  needed — owned by the consumer (the domain package), not by whatever
  eventually implements it.
- A `Service` holding business logic (validation, etc.) that depends
  only on the `Repository` interface, never a concrete database type.
- A hand-written in-memory fake implementing `Repository`, used in a
  table-driven test suite with no database required.

## Current status

**Done:**
- Frontend: browsable end-to-end against local mock data (products,
  categories, orders). Supabase auth code exists and is fully
  functional but currently disconnected by design, so every screen is
  reachable without signing in during frontend development.
- Backend: `listing`, `category`, and `order` domain types,
  `Repository`/`Service` layers, and full test suites — all passing.
- First two SQL migrations (`categories`, `listings` +
  `listing_images`).

**In progress:**
- Wiring `pgx` to an actual Postgres database and writing real
  `Repository` implementations (currently only in-memory fakes exist
  — nothing yet reads or writes real data).

**Not started:**
- `User` domain type.
- JWT verification against Supabase (`internal/auth`).
- HTTP routing/middleware (`net/http`, Go 1.22 pattern routing).
- Reconnecting the frontend to a live backend instead of mock data.
- `cmd/server/main.go` — the actual running server.

In short: this is an active work-in-progress, not a finished product.
The backend has no live server yet — what exists is the domain layer
(types, business logic, tests) built with the intent of being
persistence- and framework-agnostic, which is why it was built in that
order.
