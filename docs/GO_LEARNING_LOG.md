# Go Backend Learning Log

This tracks progress through learning Go by building Bazar's backend
by hand. If this session is lost (PC shutdown, new chat, etc.), read
this file first to see exactly where things stand.

## How we're working (read this first, especially if you're a fresh Claude session)

The user is building this backend themselves to genuinely learn Go and
general engineering practice — not having Claude write it for them.
Claude's role is: explain one concept, assign a small concrete task,
review what the user writes, give specific itemized feedback (what's
wrong, why, and a question or hint where it's a design decision rather
than a hard syntax rule) — **without rewriting the file for them**.
Repeat until a piece is correct, then move to the next concept.

This started because Claude initially wrote the entire Go backend in
one pass after the project's structure plan was approved, and the user
pushed back hard ("why did you write everything I wanted to learn").
That commit was reverted (`git reset --hard`, local-only, never
pushed). Don't repeat that mistake — see the memory entry
`teach-dont-implement-learning-code` for the full context.

Small exception worth knowing about: Claude accidentally deleted the
user's own hand-written `migrations/initializingFirstTable.sql` while
building (and later reverting) that first pass. It was never committed
to git, so it isn't recoverable. Its content, in case it's useful as a
starting point when we get to migrations:
```sql
create table categories(
    id serial primary key,
    name text not null,
    slug text unique not null,
    image_url text nit null
)
```
(Note the original had a typo — `nit null` instead of `not null`.)

## Overall learning path (Go side)

From the project's approved plan, roughly in order:
1. Module basics + why `internal/` is special — **done**
2. Domain types (plain structs, no framework annotations) — **done**
   (`Listing`, `Category`; `Order`/`User` domain types not started yet)
3. `Repository`/`Service` interface pattern — **done**
4. Table-driven unit tests against an in-memory fake repository — **in progress**
5. `pgx` + `golang-migrate` for real persistence — not started
6. Stdlib `net/http` (Go 1.22 pattern routing) + hand-rolled middleware — not started
7. `internal/auth` — verifying Supabase JWTs — not started
8. `internal/config`, `internal/platform/{logger,database,middleware}` — not started
9. Wiring it all together in `cmd/server/main.go` — not started

Only after the Go backend has a real, working vertical slice do we move
to the separately-agreed `web/` frontend fixes (path aliases, moving
mock data out of `assets/`, naming, env hygiene) — that part is NOT
learning-gated the same way; Claude can implement it directly when we
get there, unless the user says otherwise.

## Detailed progress

### ✅ Lesson 1 — modules & `internal/`
- `go.mod`: `module github.com/kalin-roman/Bazar`, `go 1.22.1`.
- `cmd/server/main.go`: exists with just `package main` — deliberately
  has no `func main()` yet, will get built up once a real feature
  exists to wire in.
- `internal/` directory created.

### ✅ Lesson 2 — domain types
`internal/listing/listing.go`:
```go
package listing

type Listing struct {
	ID           int64
	CategoryID   int64 // foreign key to link to the categories
	Title        string
	Slug         string
	ImagesURL    []string
	PriceCents   int64
	HeroImageURL string
	MaxQuantity  int64
}
```
`internal/category/category.go`:
```go
package category

type Category struct {
	ID       int64
	Name     string
	ImageURL string
	Slug     string
}
```
Concepts covered: exported vs. unexported fields (capitalization is
the visibility mechanism); Postgres has no unsigned integer types, so
`int64` not `uint64` for IDs/money; storing price as integer cents
(`PriceCents`, ×100) instead of `float64` to avoid floating-point
rounding; Go initialism casing (`ID`, `URL`, not `Id`, `Url`); a
`Category` must never hold its listings — the reference is
one-directional (`Listing.CategoryID` → `Category.ID`), avoiding a
"god struct" / stale embedded data.

### ✅ Lesson 3 — Repository/Service pattern

`internal/listing/service.go` (final state, builds + vets clean):
```go
package listing

import (
	"context"
	"errors"
)

var ErrInvalid = errors.New("listing: invalid listing")

type Repository interface {
	List(ctx context.Context) ([]Listing, error)
	GetBySlug(ctx context.Context, slug string) (Listing, error)
	Create(ctx context.Context, l Listing) (Listing, error)
}

type Service struct {
	Repository Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repository: repo}
}

func (s *Service) List(ctx context.Context) ([]Listing, error) {
	return s.Repository.List(ctx)
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (Listing, error) {
	return s.Repository.GetBySlug(ctx, slug)
}

func (s *Service) Create(ctx context.Context, l Listing) (Listing, error) {
	if l.CategoryID == 0 || l.MaxQuantity < 0 || l.PriceCents < 0 {
		return Listing{}, ErrInvalid
	}
	if l.HeroImageURL == "" || len(l.ImagesURL) == 0 || l.Slug == "" || l.Title == "" {
		return Listing{}, ErrInvalid
	}
	return s.Repository.Create(ctx, l)
}
```

Note on how this landed: after several rounds of itemized review (see
git history / prior session for the full back-and-forth — unnamed
`ctx` params, `.append()` called as a method instead of the builtin,
slice-vs-string `==` comparisons, a bare `_ErrInvalid` token, `new(repo)`
called on a value instead of `&Service{Repository: repo}`), the user
explicitly asked Claude to write the final fix directly rather than
continue iterating via hints — a deliberate one-off exception to the
teach-don't-implement approach for this file, not a change to the
overall policy (see `teach-dont-implement-learning-code` memory).
Resume hint-only mode by default from Lesson 4 onward unless asked
otherwise again.

Design decisions resolved along the way:
- `GetBySlug` returns a single `Listing`, not `[]Listing` — a slug
  uniquely identifies one row.
- `List` stays `[]Listing` — it's the one method that's actually
  meant to return many.
- `Create` does **not** validate `l.ID` — at the point `Create` is
  called nothing has been persisted yet, so the caller has no
  meaningful ID to supply; assigning one is the repository/DB's job.
- One sentinel `ErrInvalid` covers both validation branches (checked
  via `errors.Is`), replacing the earlier typo'd raw `errors.New(...)`
  strings. Distinguishing *which* field failed (e.g. via
  `fmt.Errorf("%w: ...", ErrInvalid)`) was flagged as a possible later
  refinement, not done yet.

### 🔧 Lesson 4 — table-driven tests against an in-memory fake `Repository` (IN PROGRESS)

Not started yet beyond this point. Goal: `internal/listing/listing_test.go`,
testing `Service` (List/GetBySlug/Create) against a hand-written fake
that implements the `Repository` interface in-memory — this is the
actual payoff of having split `Repository` out as an interface in
Lesson 3. Teaching mode (hints + review, user writes the code) applies
here by default.

### Not started yet
- `internal/category` gets the same `Repository`/`Service` treatment.
- `internal/order` — bigger: has line items, a price snapshot at
  purchase time, and always-belongs-to-a-user semantics.
- `internal/auth`, `internal/platform/*`, `internal/config`,
  `cmd/server/main.go` wiring.
