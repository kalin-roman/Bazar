-- Reverse of up.sql: convert back to integer first, since the FK
-- constraint below can't be added while the column is still text.
alter table orders alter column user_id type integer using user_id::integer;

-- Re-add the exact constraint dropped in up.sql, restoring the
-- original schema.
alter table orders add constraint orders_user_id_fkey foreign key (user_id) references app_users(id);