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
   (`Listing`, `Category`, `Order`, `User`)
3. `Repository`/`Service` interface pattern — **done** (all four
   packages)
4. Table-driven unit tests against an in-memory fake repository —
   **done** (all four packages)
5. `pgx` + `golang-migrate` for real persistence — **in progress**
   (schema migrations done + validated against real Postgres; `pgx`
   connection + real `Repository` implementations not started)
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

### 🔧 `pgx` + `golang-migrate` — real persistence (IN PROGRESS)

Decision made along the way: no ORM (Sequelize doesn't apply — that's
JS/TS only; the Go equivalent would be GORM). Sticking with raw `pgx`
+ hand-written SQL deliberately, since the `Repository` interface
already makes this swappable later if ever wanted, and an ORM would
hide exactly the SQL/schema mechanics this stage is meant to teach.

`migrations/000001_create_categories.up.sql` / `.down.sql` — fully
user-written over several review rounds (braces vs. parens, `DELETE`
vs. `DROP TABLE`, a stray empty `()`, missing semicolons — all
mechanical, not conceptual):
```sql
create table categories(
    id serial primary key,
    name text not null,
    slug text unique not null,
    image_url text not null
);
```
```sql
drop table categories;
```
Note: `up.sql` was accidentally deleted from disk at one point (cause
unknown — not from any `rm` Claude ran; the `migrations/` directory
isn't tracked in git yet so there was no backup) and recreated from
this log's own record of the last-confirmed-correct content. Worth
committing `migrations/` to git soon so this can't happen silently
again — offered, not yet done at user's request.

`migrations/000002_create_listings.up.sql` / `.down.sql` — first pass
fully user-written (many issues: missing commas, `serial` used for
non-PK columns, invalid inline `foreign key` syntax, camelCase column
names, an `image_url` column left on `listings` despite deciding on a
separate table), then Claude rewrote it directly (asked and confirmed
— one-off exception, see `teach-dont-implement-learning-code` memory).
Two tables per the earlier images-storage decision:
```sql
create table listings(
    id serial primary key,
    category_id integer not null references categories(id),
    title text not null,
    slug text unique not null,
    price_cents integer not null,
    hero_image_url text not null,
    max_quantity integer not null
);

create table listing_images(
    id serial primary key,
    listing_id integer not null references listings(id),
    url text not null,
    position integer not null
);
```
`down.sql` drops `listing_images` before `listings` (FK order).

Neither migration has been validated against a real running Postgres
yet (Docker daemon wasn't up when tried) — reviewed by eye only.
`golang-migrate` CLI itself isn't installed (`migrate` not on PATH).

### ✅ `internal/user` — domain type + Repository/Service

Came up because `orders`' migration needs a `user_id` FK, but `User`
didn't exist yet — user built this package proactively to unblock
that (not formally assigned as a lesson first).

Real design question that came up along the way: the plan (from
before this session, item 7) already called for `internal/auth`
verifying **Supabase JWTs** — meaning auth is delegated to Supabase,
while the actual application data lives in Postgres via `pgx`. Those
aren't in conflict (a JWT issuer and a data store can be two
different things), but it settles an important point: this backend
does **not** own credentials. Decision: keep Supabase for auth, no
password storage in `internal/user` at all.

`internal/user/user.go` (final, after removing a first-draft
`Password string` field, fixing an accidentally-unexported
`fullName`, and renaming a copy-pasted `HeroImageURL` to `Avatar`):
```go
package user

type User struct {
	ID              int64 // For future instead of the ID I should change it on the UUID from the Supabase
	FullName        string
	Email           string
	Age             int64
	AddressDelivery string
	Avatar          string
}
```
`internal/user/service.go` — same `Repository`/`Service` shape as the
other three packages (`List`, `GetByID`, `Create`), `Create` validates
`FullName`/`Email` non-empty (no `ID == 0` check — same resolved
question as `listing`/`category`). Builds and vets clean. Not tested
yet (`service_test.go` not written).

Also created: `PORTFOLIO.md` at repo root — a recruiter-facing
technical overview (architecture, what's demonstrated, honest current
status), separate from the existing frontend-focused `README.md`.
Claude wrote it directly (explicitly not learning-gated — same as the
`web/` frontend fixes). User intends to write/overwrite `README.md`
personally later; Claude is not touching that file.

### ✅ `internal/user` — table-driven tests against an in-memory fake

`internal/user/service_test.go` — `fakeRepository` (in-memory
`[]User`), `TestServiceList`, `TestServiceGetByID`, `TestServiceCreate`
(valid + missing `FullName` + missing `Email`). All passing.

Claude wrote this directly (explicit, unambiguous request this time —
"do the test instead of me" — one-off exception, see
`teach-dont-implement-learning-code` memory).

All four domain packages (`listing`, `category`, `order`, `user`) are
now complete: types, `Repository`/`Service`, tests, all passing. User
is now working on the `users`/`orders`/`order_items` migrations in
parallel with this.

### ✅ `users` table migration

`migrations/000003_create_users.up.sql` / `.down.sql` — fully
user-written, many review rounds (the most of any migration so far —
worth noting for pattern, not as a concern): a wrong-model attempt at
a single catch-all "drop everything" down file instead of a proper
paired `000003` down migration; several casing round-trips (`fullName`
→ `FullName` → `full_name`, i.e. went further wrong before landing
right); a `users_email`/`users_password` table-prefix habit that
needed correcting twice (once for email, then password separately);
the table name itself flip-flopping `users` → `user` → `users` across
rounds, independently of the filename being correct sooner; `id`
staying `integer` instead of `serial` for several rounds after being
flagged. All converged in the end without Claude writing any of it:
```sql
create table users(
    id serial primary key,
    full_name text not null,
    email text not null,
    age integer not null,
    address_delivery text not null,
    avatar text not null
);
```
```sql
drop table users;
```
Note: no `password` column, consistent with the Supabase-owns-auth
decision — matches `internal/user`'s Go struct.

### ✅ `orders`/`order_items` table migrations

`migrations/000004_create_orders.up.sql` / `.down.sql` — Claude wrote
directly (asked and confirmed, one-off exception). `orders` (`id`,
`user_id` FK → `users`) and `order_items` (`id`, `order_id` FK →
`orders`, `listing_id` FK → `listings`, `price_cents`, `quantity`) in
one migration, `down.sql` drops `order_items` before `orders` (FK
order). Note: SQL column is `price_cents`, matching `listings`'
convention, even though the Go `OrderItem.Price` field was never
renamed to match (flagged earlier, still an open/deferred nit).

All four schema migrations are done: `categories`, `listings` +
`listing_images`, `users`, `orders` + `order_items`. None validated
against a real running Postgres yet (Docker wasn't up when tried
earlier) — reviewed by eye only.

### ✅ `golang-migrate` CLI — installed and validated against real Postgres

`migrate` CLI installed via `go install
github.com/golang-migrate/migrate/v4/cmd/migrate@latest` (in
`~/go/bin`). Postgres: an existing Docker container named `Bazar`
(`postgres` image, port `5431` on the host, password `bazarOnGo`) —
was stopped, started it. Note there's a stray pre-existing `example`
table in that database unrelated to this project's schema, left alone.

Ran `migrate ... up` (all 4 applied in FK-dependency order), inspected
the resulting schema directly via `psql \d` on every table — matched
what was reviewed exactly (columns, types, FKs) — then `migrate ...
down -all` (all 4 rolled back cleanly in reverse order, no FK errors),
confirmed the database was back to empty (only the unrelated `example`
table + `migrate`'s own `schema_migrations` tracking table remained),
then re-applied `up` to leave the container in a useful state. Full
round-trip validation, both directions, against a real database — not
just reviewed by eye anymore.

This step was infra/tooling (installing a CLI, running commands
against a database), not learning-gated Go/SQL writing, so Claude did
it directly without asking — consistent with how `go build`/`go
vet`/`go test` have been run directly throughout, as distinct from
writing the actual Go or SQL.

### ✅ `pgx` wiring — `db/db.go`

```go
package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInvalid = errors.New("db: invalid connection pool")

func New(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		wrapped := fmt.Errorf("%w: %v", ErrInvalid, err)
		return nil, wrapped
	}
	ping := pool.Ping(ctx)
	return pool, ping
}
```
Fully user-written, several review rounds. Real bugs along the way,
not just style: an early attempt returned `pool.Ping` (the method
value, uncalled, from an undefined `pool` variable — the actual
variable was named `res`) instead of calling it; `pgxpool.New`'s own
error went unchecked for a couple of rounds, meaning a failed pool
creation would still fall through to calling `.Ping` on a bad pool;
`fmt.Errorf`'s result was computed and then discarded (never assigned
or returned) rather than actually wrapping the returned error — caught
by `go vet` (`result of fmt.Errorf call not used`), not just review;
and one round briefly returned `&pgxpool.Pool{}` (a zero-value, totally
non-functional pool) instead of `nil` on the failure path, which
would've been worse than the plain unwrapped error it replaced. All
resolved. `pgx/v5` added via `go get` directly (package management,
not learning-gated) — bumped `go.mod`'s `go` directive from `1.22.1` to
`1.25.0` as a required side effect, flagged to user, doesn't affect the
planned Go 1.22 `net/http` pattern-routing lesson (still available in
1.25+). Not yet called from anywhere (no `main.go` wiring yet, by
design — that's a later lesson).

### ✅ `category`'s real `Repository` implementation — `internal/storage/postgres/categories.go`

New package/location decided by user (their own reasoning, sound):
`internal/storage/postgres`, one file per domain package it implements
(`categories.go` first; `listing.go`/`order.go`/`user.go` later,
mirroring the same pattern). Added `category.ErrNotFound` (exported
sentinel — previously only the test fake had an unexported one) so
real callers have something to `errors.Is` check against.

`GetBySlug` — fully user-written over several review rounds: an
`ID`/`ImageURL` receiver-name confusion early on that turned out to be
"`Repository` doesn't exist yet, you have to define it" rather than an
import problem; a genuine SQL bug (`from the categories` — extra word,
confirmed by actually running the exact query against the real
`Bazar` Postgres container and getting `ERROR: relation "the" does not
exist`) caught before Claude touched the file; `fmt.Println(...,
"%s", err)` misusing `Println` for `Printf`-style formatting, caught
by `go vet` directly. User then asked Claude to review-and-fix
directly (one-off exception): removed the print entirely (a repository
method logging its own errors was flagged as mixing concerns), added
the `pgx.ErrNoRows` → `category.ErrNotFound` translation, switched to
pointer receiver, added `var _ category.Repository = (*Repository)(nil)`.

That interface-satisfaction check caused user confusion for a few
messages — with only `GetBySlug` implemented, `*Repository` correctly
failed to satisfy `category.Repository` ("missing method Create"),
which read as broken code rather than the check doing its job (it's
supposed to fail until all three methods exist). Worth remembering for
future sessions: introduce this check with an explicit heads-up that
it errors until the interface is fully implemented, not just as a
one-line aside.

`List`/`Create`: user's first attempt at `List` had real syntax errors
(`row.)`, an empty `for range`, `QueryRow` used with a `List` query
that still referenced a `slug` parameter that didn't exist) — user then
asked Claude to implement both directly (one-off exception). Final
version validated against the real `Bazar` Postgres container end to
end, not just compiled: a temporary throwaway `cmd/verify-temp/main.go`
exercised `Create` → `GetBySlug` (found) → `GetBySlug` (missing, got
`category.ErrNotFound`) → `List` → cleanup, all correct, then deleted
(never meant to be committed).

`category`'s full stack is now done: domain type, `Repository`/
`Service`, fakes/tests, and a real `pgx`-backed implementation,
validated against real Postgres.

### Not started yet

Remaining core lessons, roughly in order:
1. Real `Repository` implementations for the other three domain
   packages (`listing`, `order`, `user`), same `internal/storage/postgres`
   pattern as `category` — `listing.go`, `order.go`, `user.go`.
2. `net/http` routing (Go 1.22 pattern routing) + hand-rolled
   middleware.
3. `internal/auth` — verifying Supabase JWTs.
4. `internal/config`, `internal/platform/{logger,database,middleware}`.
5. `cmd/server/main.go` — wiring everything together into an actual
   running server.

Deferred/optional polish, not blocking, revisit if relevant later:
- `Order.Total()` method (sum `Price * Quantity` across `Items`).
- Reconciling `User.ID` (`int64`) with Supabase's UUID-based identity.
- Renaming `OrderItem.Price` → `PriceCents`, matching the SQL column
  and the rest of the codebase's naming convention.
