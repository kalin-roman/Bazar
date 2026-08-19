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
3. `Repository`/`Service` interface pattern — **in progress**
4. Table-driven unit tests against an in-memory fake repository — not started
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

### 🔧 Lesson 3 — Repository/Service pattern (IN PROGRESS)

Working file: `internal/listing/service.go`.

Concepts covered so far: interfaces in Go are implicit/structural (no
`implements` keyword); the *consumer* defines the interface it needs
(`Repository` lives in `listing`, even before any implementation
exists); `Service` holds the `Repository` interface, not a concrete
type, so it can be unit-tested with an in-memory fake later;
`context.Context` as the idiomatic first parameter for I/O-touching
methods; sentinel errors (`var ErrInvalid = errors.New(...)`) checked
via `errors.Is`, instead of matching raw strings.

**Open issues as of the last review** (fix these, then show the file again):
1. `GetBySlug` has no function body — compile error.
2. `NewService` has no function body — compile error.
3. `Create` still builds `new([]Listing)` and calls `.append(...)` as
   if it were a method (it's a built-in function, `append(slice, item)`,
   not `slice.append(...)`) — but more importantly, `Create` shouldn't
   build a slice at all. It should validate, then delegate to
   `s.Repository.Create(ctx, l)` and return the single `Listing` that
   comes back (mirroring the pattern below).
4. `l.ImagesURL == ""` doesn't compile — slices can't be compared with
   `==` to a string. Use `len(l.ImagesURL) == 0` to check emptiness.
5. `l.ID < 0` is being validated in `Create`, but open question for the
   user to resolve: at the moment `Create` is called, has this listing
   been saved anywhere yet? Where would a meaningful `ID` value even
   come from at that point?
6. No `ErrInvalid` sentinel defined yet — still raw `errors.New("...")`
   strings with typos in the messages.
7. `GetBySlug` returns `[]Listing` (a slice) in both the `Repository`
   interface and the `Service` method. Open question: a slug uniquely
   identifies one listing — should this return a single `Listing`
   instead?
8. `Repository`'s `GetBySlug` and `Create` methods are missing
   `ctx context.Context` as their first parameter — `List` already has
   it; all three should be consistent.

Worked example already given (to generalize the pattern from, not to
copy elsewhere without understanding it) — this fixes two syntax
mistakes: an unnamed parameter (so there was no `ctx` variable to
reference inside the body) and calling-syntax vs. declaration-syntax
confusion:
```go
func (s *Service) List(ctx context.Context) ([]Listing, error) {
	return s.Repository.List(ctx)
}
```

### Not started yet
- `internal/listing/listing_test.go` — table-driven tests against an
  in-memory fake `Repository` (the actual payoff of the interface
  split).
- `internal/category` gets the same `Repository`/`Service` treatment.
- `internal/order` — bigger: has line items, a price snapshot at
  purchase time, and always-belongs-to-a-user semantics.
- `internal/auth`, `internal/platform/*`, `internal/config`,
  `cmd/server/main.go` wiring.
