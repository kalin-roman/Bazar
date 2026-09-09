-- users.id will hold Supabase's UUID (caller-supplied) instead of a
-- Postgres-generated serial value, so the auto-increment default has
-- to go before the column's type can change — otherwise it would keep
-- trying to call the integer sequence's nextval() on a text column.
alter table users alter column id drop default;

-- Existing rows already hold integer values; ::text tells Postgres
-- how to convert each one (e.g. 5 -> '5') to the new type. No
-- constraint to drop first this time (unlike orders.user_id) — the
-- primary key constraint doesn't care what type the column holds.
alter table users alter column id type text using id::text;
