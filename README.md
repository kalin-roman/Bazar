# Bazar

[![Go](https://github.com/kalin-roman/Bazar/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/kalin-roman/Bazar/actions/workflows/go.yml)

A full-stack mobile marketplace: a Go + PostgreSQL backend (deployed,
live) paired with a React Native / Expo frontend, connected end to
end — real Supabase auth, real checkout, real data, no mocks.

**Live API**: [bazar-xxjl.onrender.com](https://bazar-xxjl.onrender.com)
(every route requires a valid Supabase token, so a bare visit
correctly returns `401`, not a page).

For the full technical story — architecture decisions, what this
project demonstrates, and an honest account of what's still open —
see [`PORTFOLIO.md`](PORTFOLIO.md). This file covers how to actually
run it.

## Project layout

This is two projects in one repository:

```
Bazar/
├── cmd/, internal/, migrations/, db/   # Go backend (this directory)
├── web/                                 # React Native / Expo frontend
├── PORTFOLIO.md                          # Full technical write-up
└── docs/GO_LEARNING_LOG.md                # Lesson-by-lesson backend build log
```

The backend and frontend are run and set up independently — pick the
section below for what you're working on.

---

## Backend (Go)

### Prerequisites

- [Go](https://go.dev) 1.25+
- [Docker](https://www.docker.com) (for a local Postgres instance)
- [`golang-migrate`](https://github.com/golang-migrate/migrate) CLI:
  `go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest`

### Run it locally

1. **Start a local Postgres**, e.g.:
   ```bash
   docker run -d --name bazar-postgres -p 5431:5432 -e POSTGRES_PASSWORD=<your-password> postgres
   ```

2. **Apply migrations:**
   ```bash
   export DATABASE_URL="postgres://postgres:<your-password>@localhost:5431/postgres"
   migrate -database "$DATABASE_URL?sslmode=disable" -path migrations up
   ```

3. **Set the remaining required environment variables:**
   - `DATABASE_URL` — as above.
   - `JWKS_URL` — a Supabase project's JWKS endpoint, e.g.
     `https://<your-project-ref>.supabase.co/auth/v1/.well-known/jwks.json`.
     The backend verifies incoming JWTs against this; it doesn't issue
     tokens itself (that's Supabase Auth's job, wired up on the
     frontend side).

4. **Run the server:**
   ```bash
   go run ./cmd/server
   ```
   Listens on `:8080`. Every route requires `Authorization: Bearer
   <token>` — a valid Supabase-issued JWT for the project named in
   `JWKS_URL`.

5. **(Optional) Seed some catalog data** to browse, since a fresh
   database starts empty:
   ```bash
   go run ./cmd/seed-temp
   ```

### Tests

```bash
go build ./...
go vet ./...
gofmt -l .        # should print nothing
go test ./...
```
Same checks CI runs on every push (see the badge above).

### Deploying

The backend ships as a Docker image (`Dockerfile` at the repo root) —
currently deployed on [Render](https://render.com). The binary always
listens on `:8080` (hardcoded, not read from an env var); the only
environment variables it actually requires are `DATABASE_URL` (a
Postgres connection string — the deployed instance uses Supabase's
session pooler) and `JWKS_URL`.

---

## Frontend (React Native / Expo)

All frontend commands run from the `web/` directory.

### Prerequisites

- [Node.js](https://nodejs.org) (v18+)
- [Expo Go](https://expo.dev/client) app, for testing on a physical
  device (optional — simulators/web work too)

### Setup

```bash
cd web
npm install
```
`npm install` needs the committed `.npmrc` (`legacy-peer-deps=true`)
to succeed — it's already in place, no action needed, just don't
remove it.

### Run it

```bash
npm start          # Expo dev server — press i/a/w, or scan the QR code
npm run ios         # iOS simulator directly
npm run android      # Android emulator directly
npm run web           # Browser
```

By default this talks to the **live production API**
(`bazar-xxjl.onrender.com`) and a **live Supabase project** — sign up
or sign in with a real email/password on first launch. To point it at
a different backend, change `API_BASE_URL` in `web/src/lib/api.ts`.

### Type-checking

```bash
npx tsc --noEmit
```
No dedicated lint/test script is configured for the frontend.

---

## Learning context

This backend was built by hand, from scratch, as a structured project
to learn Go and backend engineering practice — not generated in one
pass. [`docs/GO_LEARNING_LOG.md`](docs/GO_LEARNING_LOG.md) is the
real, session-by-session record of that process: what was taught,
what broke, and why — kept up throughout rather than written after
the fact.

## License

This project is private and not licensed for public distribution.
