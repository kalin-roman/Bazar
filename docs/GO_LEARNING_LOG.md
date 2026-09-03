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

### ✅ `listing`'s real `Repository` implementation — `internal/storage/postgres/listing.go`

Added `listing.ErrNotFound` (same reasoning as `category`'s). First
draft (user-written) hit the same type-collision issue `category`'s
constructor pattern implied but didn't warn about explicitly: reused
`Repository` (the type from `categories.go`) for `listing`'s methods
too — `method Repository.List already declared`, since a Go type can
only have one method of a given name regardless of differing
signatures. Fixed by introducing a separate `ListingRepository` type
(worth remembering: **future domain repositories need their own named
struct type, not a shared `Repository`** — flag this explicitly next
time before assigning the task, don't just leave it as an implicit
"decide the shape" question).

Beyond that, first draft had real runtime-only bugs invisible to
`go build`/`go vet` (SQL is just a string to the compiler): `Scan`
called with bare values instead of `&`-prefixed pointers (compiles —
`Scan` takes `...any` — but pgx errors at runtime on non-pointer
destinations); `Scan` destination order not matching the `select`'s
column order; `insert into listing` (singular, wrong table name);
an insert column-list/placeholder-count mismatch (6 columns, 7
placeholders); attempting to insert/scan `l.ID` and `l.ImagesURL` where
neither belongs (`ID` is the auto-generated PK, `ImagesURL` lives in
the separate `listing_images` table, not a `listings` column). User
asked Claude to fix directly (one-off exception). `ImagesURL` is
deliberately left unpopulated (`nil`) by every method here — treated
as explicitly deferred (needs a second query or join against
`listing_images`, real added work) rather than guessed at.

Validated against the real `Bazar` Postgres container end-to-end via a
temporary `cmd/verify-temp/main.go` (seeded a category to satisfy the
FK, then `Create` → `GetBySlug` found → `GetBySlug` missing →
`listing.ErrNotFound` → `List` → cleanup), deleted after.

### ✅ `order`'s real `Repository` implementation — `internal/storage/postgres/order.go`

Added `order.ErrNotFound`, same pattern as the other three. This one's
structurally different from `category`/`listing`: `Order.Items` lives
in a separate `order_items` table, so every method needs a second
query. Added a private `itemsForOrder` helper (fetches `order_items`
by `order_id`) that both `GetByID` and `List` call to assemble the
full `Order` value — `List` calls it once per order in a loop (the
simple, more-round-trips approach, flagged as fine for now over a
join).

`Create` introduces **transactions** — new concept, first time it's
come up. Writing an `Order` means two tables (`orders`, then one
`order_items` row per item); a transaction (`pool.Begin(ctx)` → do all
writes against the returned `tx`, not the pool directly → `tx.Commit`
only if everything succeeded, with `defer tx.Rollback(ctx)` as a
safe-if-already-committed no-op backstop) is what actually answers the
"what if an insert fails partway through" question raised when this
task was assigned — without it, a failed `order_items` insert could
leave a saved `orders` row with missing/partial items behind.

User asked Claude to write this file directly (one-off exception) —
wrote the struct/constructor/interface-check shape correctly
themselves first (visible progress: no longer needing that pattern
re-explained). Validated end-to-end against the real `Bazar` Postgres
container (seeded a user + category + listing to satisfy FKs, then
`Create` with 2 line items → `GetByID` found, both items present →
`GetByID` missing → `order.ErrNotFound` → `List` → cleanup).

Three of the four domain packages now have complete, validated real
`pgx` implementations: `category`, `listing`, `order`. **`user` still
pending** — smallest of the four in scope, but worth remembering it
still needs its own `UserRepository` type + constructor, following the
exact same pattern.

### ✅ `user`'s real `Repository` implementation — `internal/storage/postgres/user.go`

Added `user.ErrNotFound`, same pattern as the other three. Single
table (`users`), no multi-query complexity like `order`'s — a clean,
uneventful pass mirroring `category`'s shape exactly. Claude wrote it
directly (asked and confirmed). Validated end-to-end against the real
`Bazar` Postgres container (`Create` → `GetByID` found → `GetByID`
missing → `user.ErrNotFound` → `List` → cleanup).

**Persistence lesson (item 5 on the original roadmap) is now fully
complete**: all four domain packages (`category`, `listing`, `order`,
`user`) have types, `Repository`/`Service`, fake-backed tests, and
real, validated `pgx` implementations in `internal/storage/postgres`.

### 🔧 `net/http` routing — first endpoint done, `category` only so far

Layering established: HTTP handler → `Service` → `Repository` (each
layer knows only the one below it — handler never touches `pgxpool`,
mirrors the domain/persistence split from earlier lessons). New
top-level layout: `internal/http/<domain>http/` per domain package,
mirroring `internal/storage/postgres/`'s one-file-per-domain pattern —
so far just `internal/http/categoryhttp/`.

`handlers.go` (fully user-written, several rounds): `HandlesService`
struct holding `*category.Service`, `NewCategoriesService` constructor,
`List` method. Real conceptual mixups along the way, not just syntax:
package named `category` colliding with the imported domain package of
the same name (fixed → `categoryhttp`); `New`'s body calling
`category.NewService(s)` on an already-constructed `*Service` instead
of constructing this file's own handler struct; a `var _
category.Service = (*HandlesService)(nil)` interface-check attempted
against `category.Service`, which is a concrete struct, not an
interface (that check only means something against an interface type
— removed). Also: user asked "should this use pgxpool.Pool" — good
question, answered no, reinforced the full layering chain explicitly.
Final `List` method correct: `r.Context()`, calls `Service.List`,
writes JSON, minimal error handling. Minor open items, not blocking:
`category` (the local var) shadows the package name, `Encode`'s error
is silently ignored — both flagged, neither fixed yet.

`router.go`: `NewRouter` builds an `*http.ServeMux`, registers `GET
/categories` → `handler.List`. Two real bugs, both compiler-caught:
`*http.ServerMux` (typo, should be `ServeMux`) and `mux.Handle(...)`
used instead of `mux.HandleFunc(...)` (`Handle` needs something that
already implements the `http.Handler` interface; a plain function
needs `HandleFunc`, which adapts it). Both fixed by user.

Validated end-to-end for real: a temporary harness wired real
Postgres → `CategoryRepository` → `category.Service` → `HandlesService`
→ `NewRouter`, started an actual `http.Server`, and `curl`'d
`GET /categories` — got `200`, body `null` (table's empty right now,
which is correct — `nil` slice encodes as JSON `null`, not `[]`; noted
as an open API-design decision, not a bug, for later). Deleted after.

Incidental cleanup along the way: `categories.go`'s constructor got
renamed by the user to match the other three's naming convention
(`CategoryRepository`/`NewCategorieRepository` — small typo, missing
an `s`, `NewCategoryRepository` was probably intended).

### ✅ `listing`'s HTTP handler + router — `internal/http/listinghttp/`

Fully user-written, went noticeably faster than `category`'s first
pass — mostly mechanical adaptation this time rather than new
concepts. Real thing that came up along the way: a design question
about consolidating routing. User asked whether all domains' routes
could live in one common router file instead of each domain building
its own separate `*http.ServeMux`. Answer: yes, and it's actually the
more correct shape once there's more than one domain — an `http.Server`
needs exactly one handler, so four independent muxes would need
merging (fiddly) versus one shared mux with every domain's routes
registered onto it (simpler). Decided to defer building that
consolidation until the `cmd/server/main.go` wiring lesson, since
that's the same job (assembling every piece together) — each domain's
`*http` package keeps its own `NewRouter` for now, still useful
standalone.

`handlers.go`: first draft was copy-pasted from `categoryhttp` without
fully adapting names — constructor still called `NewCategoriesService`
inside `listinghttp`, and the `List` method's result variable was
still named `category` despite holding `[]listing.Listing`. Both
fixed by user (→ `NewListingService`, → `listing` — though that still
shadows the imported `listing` package for the rest of the function,
same low-priority habit flagged before, harmless here, not fixed).

`router.go`: correct on `ServeMux`/`HandleFunc` this time (no repeat
of `categoryhttp`'s typos) — only issue was the registered path being
singular (`/listing`) instead of plural (`/listings`), fixed.

Validated end-to-end same as `category`: temporary harness wired real
Postgres → `ListingRepository` → `listing.Service` → handler → router,
real `http.Server`, `curl GET /listings` → `200`, `null` (empty table,
expected). Deleted after.

Two domains done (`category`, `listing`), two left (`order`, `user`)
for handler/router — then hand-rolled middleware, not started yet.

### ✅ `order`'s HTTP handler + router — `internal/http/orderhttp/`

Claude wrote directly (asked and confirmed). Same shape as `category`/
`listing`: `HandlesService` holding `*order.Service`,
`NewOrderService` constructor, `List` method, `NewRouter` registering
`GET /orders`. Validated end-to-end against real Postgres (`200`,
`null` — empty table).

Three domains done (`category`, `listing`, `order`), one left (`user`)
for handler/router — then hand-rolled middleware.

### ✅ `user`'s HTTP handler + router — `internal/http/userhttp/`

Claude wrote directly (asked and confirmed — user said it's "exactly
the same" as the other three, one-off exception). Same shape as
`category`/`listing`/`order`: `HandlesService` holding `*user.Service`
(field named `UserService`, matching `order`'s `OrderService`
convention rather than `category`'s `CatService`), `NewUserService`
constructor, `List` method, `NewRouter` registering `GET /users`.
`go build ./internal/...` and `go vet ./internal/...` both clean.

Not validated live against real Postgres this time — Docker Desktop's
daemon wasn't running (`docker` CLI targets its socket specifically;
the system `dockerd` was active but permission-denied on the fallback
socket) and the user opted to skip rather than start it, since
build+vet already pass and this is a mechanical mirror of three
already-validated implementations. Worth doing the live curl check
next time Docker Desktop happens to be up anyway, just to close it
out.

All four domains (`category`, `listing`, `order`, `user`) now have
complete HTTP handler + router layers.

### ⤺ Router consolidation — tried, then reverted

User's own idea, prompted by not liking 4 files with identical
structure scattered one per domain package. Built
`internal/http/router/router.go` (package `router`, named to avoid
colliding with `net/http`) holding `NewRouterCategory`/
`NewRouterListing`/`NewRouterOrder`/`NewRouterUser`, each still
building and returning its own separate `*http.ServeMux` — Claude
wrote it directly (explicit request), built + vetted clean.

On reflection (prompted by asking "is this a good approach"), flagged
a real gap: this only reorganizes files, it doesn't solve the
composability problem already named during the `listing` lesson — an
`http.Server` needs exactly one handler, and 4 independently-built
muxes still don't merge into one without either switching to a
"register onto a shared mux passed in as an argument" shape, or path
prefixing + `http.StripPrefix` (the "fiddly" option already dismissed
earlier). So this change made source layout tidier without actually
moving the real problem forward.

**Reverted** at user's request — deleted `internal/http/router/`,
recreated the original 4 `router.go` files (one per `<domain>http`
package, each with its own `NewRouter(h *HandlesService)
*http.ServeMux`), byte-for-byte the same shape as before the
consolidation attempt. Builds + vets clean. Back to the state after
`user`'s HTTP handler + router lesson.

The underlying composability question is still open and still
deferred to the `main.go` wiring lesson, as originally decided — worth
remembering when we get there that "each domain returns its own new
mux" is the shape that needs to change (to register-onto-shared-mux)
for a single `http.Server` to actually serve all four domains.

Next: hand-rolled middleware.

### ✅ First hand-rolled middleware — `internal/http/middleware/middleware.go`

User's first attempt (`internal/http/middleware.go`, `package http`)
hit the exact package-naming collision flagged during the router
lesson — naming a package `http` right next to a `net/http` import —
plus an incomplete function (no body). Claude wrote it directly
(explicit request), moving it into its own subpackage for the same
reason `router` got one: `internal/http/middleware/middleware.go`,
`package middleware`.

`Logging(next http.Handler) http.Handler` — the core hand-rolled
middleware pattern: wraps a handler, returns
`http.HandlerFunc(func(w, r) {...})` that does work around calling
`next.ServeHTTP(w, r)`. This one times the request (`time.Now()` /
`time.Since`) and logs method + path + duration after the wrapped
handler returns.

Verified with a throwaway `httptest`-based test (no Postgres needed
for this one) — confirmed the wrapped handler actually gets called
(guards against the classic bug: a middleware that forgets
`next.ServeHTTP` silently breaks everything past it) and its status
code passes through untouched; logged output confirmed real
(`GET /categories 1.203µs`). Test file deleted after, same
throwaway-harness pattern as the Postgres verification steps.

Not yet wired into any router (no `main.go` yet) — first middleware
proven standalone, chaining/application decisions deferred, same as
routing.

### ✅ `internal/auth` — verifying Supabase JWTs

Confirmed with the user first: this Supabase project uses legacy HS256
(shared-secret) signing, not the newer asymmetric ES256/RS256+JWKS —
settles which library/approach applies. Added
`github.com/golang-jwt/jwt/v5` via `go get` directly (package
management, not learning-gated, same as `pgx` earlier) — no `go.mod`
side effects this time.

`internal/auth/auth.go`: `ErrInvalidToken` sentinel (same
wrap-with-`%w` pattern as `db.New`), `VerifyToken(tokenString, secret
string) (string, error)`. Uses `jwt.ParseWithClaims` with
`jwt.RegisteredClaims`, extracts `Subject` (the `sub` claim — Supabase
puts the user's UUID there) as the returned user ID. Secret is taken
as a plain parameter for now, not sourced from config — no
`internal/config` yet, that's a later lesson.

Claude wrote this directly (explicit request — "write it instead of
me"). Real security detail baked in, not just plumbing: the keyFunc
checks `token.Method.(*jwt.SigningMethodHMAC)` before returning the
secret, which is what actually defends against the "algorithm
confusion" JWT attack class (a forged token claiming `alg: none` or a
different algorithm, trying to bypass verification) — this isn't
optional/style, skipping it is a real vulnerability.

Verified with a throwaway table of tests (deleted after, same
temp-harness pattern used throughout): valid token → correct user ID;
expired token → rejected (library handles `exp` automatically, no
manual check needed); wrong secret → rejected; and specifically an
`alg: none` token signed via `jwt.UnsafeAllowNoneSignatureType` →
rejected, proving the algorithm-confusion guard actually works and
isn't just defensive-looking dead code. `go build ./internal/...` and
`go vet ./internal/...` both clean.

Not yet wired into a middleware (an obvious next step — an `Auth`
middleware that calls `VerifyToken` on the `Authorization: Bearer`
header and rejects/passes the request accordingly) or into `main.go` —
those are part of the wiring lesson still ahead.

### ✅ Auth middleware — `internal/http/middleware/auth.go`

Mostly user-written, several review rounds, each addressing one thing —
good example of the closure-nesting concept genuinely landing after a
few passes rather than being explained once and immediately correct:

1. First draft only had the outer `Auth(secret) func(http.Handler)`
   layer — missing the middle `func(next http.Handler) http.Handler`
   layer entirely, and tried returning an `http.HandlerFunc` straight
   from layer 1. Didn't compile (confirmed via `go build`). Explained
   via walking `Logging`'s already-working two-layer shape and mapping
   each layer's type onto what `Auth` additionally needs (secret
   captured before `next` exists) — worth remembering for future
   closure-shaped lessons: anchor new nesting concepts against a
   working example already in the codebase, not just type signatures
   in the abstract.
2. Second attempt fixed the outer return type but added `http.Handler`
   as a return type on the *innermost* function literal, breaking its
   conversion to `http.HandlerFunc` (which requires exactly
   `func(http.ResponseWriter, *http.Request)`, no return value) — user
   asked Claude to write just this structural piece directly (one-off
   exception, diff shown rather than rewriting the whole file), leaving
   the header-check logic bugs untouched/unfixed on purpose.
3. Header-prefix check went through real logic bugs, caught by review
   before any test ran: inverted condition (rejected when `"Bearer"`
   *was* present, should reject when absent), an off-by-one manual
   slice (`authorized[0:7]` compared against 6-char `"Bearer"`, always
   false), and a slice that would panic at runtime on a short header
   string — same "runtime-only bug invisible to `go build`" category as
   the earlier `pgx` `Scan`-without-`&` mistake. Fixed by switching to
   `strings.HasPrefix`/`strings.TrimPrefix` (safe on short input) and
   negating the condition.
4. `auth.VerifyToken(token, secret)`'s return values were called but
   discarded entirely for a round — compiles fine (Go allows ignoring
   return values, unlike unused local variables) but meant *any* token,
   valid or forged or expired, passed straight through with no 401 —
   the middleware's whole reason for existing silently did nothing.
   Caught by review, not the compiler.
5. Final remaining piece (context propagation + calling `next`) Claude
   wrote directly (explicit request, "fix it and check again") since it
   was the one part never actually attempted yet: `ctxKey` unexported
   int type + `userIDKey` const + `UserIDFromContext(ctx)
   (string, bool)` helper (the two-value `ctx.Value(...).(string)`
   form), and the success path — `context.WithValue` +
   `next.ServeHTTP(w, r.WithContext(ctx))`.

Verified with a throwaway `httptest` table (deleted after): no
`Authorization` header → 401, `next` not called; garbage token → 401,
`next` not called; valid signed token → `200`, `next` called, and
`UserIDFromContext` inside the wrapped handler correctly recovers the
`sub` claim (`"user-123"`) — proves the context propagation actually
works end-to-end, not just that it compiles. `go build ./internal/...`
and `go vet ./internal/...` both clean.

Auth middleware is now complete and proven standalone. Still not wired
into any router or `main.go` — that's the wiring lesson.

### ✅ `internal/config` — `internal/config/config.go`

User-written first draft: `Config` struct (`ConnectionString`,
`JWTsecret`) + a bare `Load() (*Config, error)` signature with no body
yet — didn't compile (`missing function body`), expected at that
stage. One real naming nit caught in review before the body existed:
`JWTsecret` should be `JWTSecret` — `JWT` is itself an initialism,
same casing convention from Lesson 2 (`ID`, `URL`, not `Id`/`Url`).

Claude wrote the fix + `Load`'s body directly (explicit request).
`ErrMissingEnv` sentinel (same `%w`-wrapping style as `db`/`auth`).
`Load` reads `DATABASE_URL` and `JWT_SECRET` via `os.LookupEnv` (not
`os.Getenv` — `LookupEnv`'s second return value distinguishes "unset"
from "set to empty string," which matters for failing fast on a
missing required secret rather than silently proceeding with a blank
one). Errors out via `ErrMissingEnv` if either is missing/empty.

Verified with throwaway tests (deleted after, using `t.Setenv` for
per-test env isolation): missing `DATABASE_URL` → error; missing
`JWT_SECRET` → error; both present → `Config` populated correctly.
`go build ./internal/...` and `go vet ./internal/...` both clean.

Note: `.env`-file loading (e.g. via `godotenv`) deliberately not
added — `os.LookupEnv` only reads real process environment variables,
so local runs need vars exported in the shell for now. Flagged as a
possible later addition, not required.

Not yet wired anywhere — nothing calls `Load()` yet (no `main.go`).

### Not started yet

Remaining core lessons, roughly in order:
1. Router composition refactor — change the four `NewRouterX` functions
   from "each builds its own `*http.ServeMux`" to "each registers
   routes onto a `*http.ServeMux` passed in", so `main.go` can build
   one shared mux across all four domains (resolves the deferred
   "4 separate muxes don't compose into one `http.Server`" question).
3. `cmd/server/main.go` — wiring everything together into an actual
   running server: config → `db.New` pool → all four
   repo/service/handler stacks → one shared mux → wrap in `Logging` +
   `Auth` middleware → `http.ListenAndServe`. Then smoke-test live
   against real Postgres end-to-end, same as every other piece.
4. `internal/platform/{logger,database,middleware}` — optional
   organizational polish, not functionally required.

`PORTFOLIO.md` was updated to reflect all of the above as of this
point (auth verification + middleware pattern + this auth middleware
all done; router composition, config, and `main.go` wiring still
ahead) — kept in sync with this log, not left stale.

Deferred/optional polish, not blocking, revisit if relevant later:
- `Order.Total()` method (sum `Price * Quantity` across `Items`).
- Reconciling `User.ID` (`int64`) with Supabase's UUID-based identity.
- Renaming `OrderItem.Price` → `PriceCents`, matching the SQL column
  and the rest of the codebase's naming convention.
