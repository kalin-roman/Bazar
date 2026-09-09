-- user_id will hold Supabase's UUID (a string) instead of a row from
-- this database's own users.id (an integer), so the FK relationship
-- no longer makes sense and has to go before the column's type can
-- change.
alter table orders drop constraint orders_user_id_fkey;

-- Existing rows already hold integer values; ::text tells Postgres
-- how to convert each one (e.g. 5 -> '5') to the new type.
alter table orders alter column user_id type text using user_id::text;
