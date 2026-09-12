-- Reverse of up.sql, in reverse order: convert back to integer first.
alter table app_users alter column id type integer using id::integer;

-- Restore the auto-increment default dropped in up.sql. The
-- underlying sequence (app_users_id_seq) was never touched by
-- up.sql, so it still exists and can be reattached directly.
alter table app_users alter column id set default nextval('app_users_id_seq'::regclass);
