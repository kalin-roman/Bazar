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
4. Table-driven unit tests against an in-memory fake repository — **done**
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

### ✅ Lesson 4 — table-driven tests against an in-memory fake `Repository`

`internal/listing/listing_test.go` — `fakeRepository` (in-memory
`[]Listing`) implementing `Repository`, plus `TestServiceList`,
`TestServiceGetBySlug`, and `TestServiceCreate` (table-driven over a
`mutate func(Listing) Listing` pattern covering each validation
branch). All passing (`go test ./internal/listing/...`).

Note on how this landed: same as Lesson 3's `service.go` fix — after
being asked twice to just implement it (once mid-explanation, then
again immediately after explicitly choosing teaching mode in an
AskUserQuestion — flagged as likely an accidental duplicate message,
which the user then confirmed was intentional), Claude wrote this file
directly as a one-off exception rather than continuing hint-only mode.
Not a change to the overall policy — see
`teach-dont-implement-learning-code` memory. Resume hint-only mode by
default for whatever comes next (`internal/category`'s
`Repository`/`Service` pattern, or `internal/order`) unless asked
otherwise again.

### ✅ `internal/category` — Repository/Service pattern

`internal/category/service.go` — same shape as `listing`'s: `Repository`
interface (`List`/`GetBySlug`/`Create`, `ctx` first param throughout),
`Service` + `NewService`, `Create` validates then delegates. Builds and
vets clean.

This one was fully user-written and reviewed in several small rounds —
back to hint-only mode as planned after Lesson 4's one-off exception.
Real bug caught in review: first draft validated `l.ID == 0` as
invalid in `Create`, same mistake the `ID` question in `listing` had
already resolved the other way (nothing's been persisted yet at that
point, so there's no meaningful `ID` to check) — fixed. Also took four
review rounds to land the `ErrInvalid` message text (`"category:
invalid category"` — went through a stray capital, a typo, a
leftover plural, and a stray capital letter along the way; each round
was minor and mechanical, not conceptual). Left as-is, by choice, not
bugs: `len(l.ImageURL) == 0` instead of `l.ImageURL == ""` (works fine,
`ImageURL` is a `string` not `[]string` so `len` isn't required but
isn't wrong either), and the `Create` parameter is still named `l`
(carried over from `listing`) rather than `c`.

### ✅ `internal/category` — table-driven tests against an in-memory fake

`internal/category/service_test.go` — `fakeRepository` (in-memory
`[]Category`), `TestServiceList`, `TestServiceGetBySlug`,
`TestServiceCreate` (3 validation branches: `ImageURL`, `Slug`, `Name`
empty). All passing.

Claude wrote this one directly (asked and confirmed, one-off exception
again — see `teach-dont-implement-learning-code` memory) rather than
the user writing it. `category`'s `service.go` itself was still fully
user-written/reviewed.

`internal/listing` and `internal/category` now both have complete
`Repository`/`Service` + tests. Both packages' domain model is done.

### ✅ `internal/order` — domain types

`internal/order/order.go`:
```go
package order

type Order struct {
	ID     int64
	UserID int64
	Items  []OrderItem
}

type OrderItem struct {
	ListingID int64
	Price     int64
	Quantity  int64
}
```
Fully user-written, several review rounds (each addressed one thing,
not fixed simultaneously — normal pace for this file, not a sign of
struggle): started with an empty `OrderItem`, a redundant
`AmountOfOrderItems` counter field (`len(Items)` makes it derivable —
same "don't store what you can cheaply compute" principle as the
Lesson-2 note about `Category` not holding its `Listing`s), a stray
`InStock bool` on `OrderItem` (current-state property, doesn't belong
on a historical line item), and one round where `ListingID` landed on
`Order` instead of `OrderItem` (would've meant one order = one
listing, contradicting having multiple `Items`). Final shape resolves
all four original design questions: `UserID` = ownership
(one-directional, same pattern as `Listing.CategoryID`), `Items
[]OrderItem` = multiple line items, `OrderItem.ListingID` = which
listing (a bare `int64`, not an embedded `listing.Listing` — `order`
doesn't import `listing` at all, keeps the packages decoupled),
`OrderItem.Price`/`Quantity` = frozen at purchase time, independent of
whatever the live `Listing` says later. Builds and vets clean.

Still-open nit, not fixed, user's call: `OrderItem.Price` vs.
`PriceCents` (matching `Listing.PriceCents`'s naming convention for
integer-cents fields).

### ✅ `internal/order` — Repository/Service pattern

`internal/order/service.go` — same shape as `listing`/`category`, with
two differences: `GetByID(ctx, id int64)` instead of `GetBySlug`
(orders have no slug), and `Create`'s validation loops over `o.Items`
checking each `OrderItem` (`ListingID == 0`, `Quantity <= 0` — note
`<= 0` not `< 0`, zero items ordered isn't valid either — `Price < 0`),
in addition to `UserID == 0` / `len(Items) == 0` on `Order` itself.
Builds and vets clean.

Claude wrote this one directly (asked and confirmed, one-off exception
— see `teach-dont-implement-learning-code` memory); `order.go` itself
was fully user-written/reviewed.

Deliberately not added: any `Total()`/order-sum business logic — noted
as a real design decision (would be a method on `Order` in `order.go`,
summing `Price * Quantity` across `Items`, not something `Service`
needs since it doesn't touch the repository) but left for the user to
decide on, not folded in unprompted.

### ✅ `internal/order` — table-driven tests against an in-memory fake

`internal/order/service_test.go` — `fakeRepository` (in-memory
`[]Order`, `GetByID` matches by `ID` not slug), `TestServiceList`,
`TestServiceGetByID`, `TestServiceCreate` (7 cases: valid, `UserID ==
0`, empty `Items`, then per-item `ListingID == 0` /
`Quantity <= 0`/`-1` / `Price < 0`). All passing.

Claude wrote this directly (asked and confirmed, one-off exception —
see `teach-dont-implement-learning-code` memory).

`listing`, `category`, and `order` are now all complete: domain
types + `Repository`/`Service` + tests, all user-written for the
domain-type files, mixed for the service/test files (see each
section above for exactly who wrote what).

### Not started yet
- Possible: `Order.Total()` method (sum `Price * Quantity` across
  `Items`) — still undecided, raise it again if relevant later.
- `internal/auth` — verifying Supabase JWTs.
- `pgx` + `golang-migrate` for real persistence (the actual
  `Repository` implementations — note an untracked `db/db.go` stub
  exists, currently just `package db`, not yet folded into this).
- `internal/platform/*`, `internal/config`, `cmd/server/main.go`
  wiring.
