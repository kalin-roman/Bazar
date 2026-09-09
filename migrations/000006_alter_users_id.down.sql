-- Reverse of up.sql, in reverse order: convert back to integer first.
alter table users alter column id type integer using id::integer;

-- Restore the auto-increment default dropped in up.sql. The
-- underlying sequence (users_id_seq) was never touched by up.sql, so
-- it still exists and can be reattached directly.
alter table users alter column id set default nextval('users_id_seq'::regclass);
