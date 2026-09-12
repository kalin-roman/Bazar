-- Named app_users, not users: the real Supabase project this runs
-- against already has its own users table (uuid id, FK'd to its
-- auth.users), predating this backend, tied to Supabase Auth. This
-- table is this backend's own application data, kept separate rather
-- than colliding with (or replacing) that one.
create table app_users(
    id serial primary key,
    full_name text not null,
    email text not null,
    age integer not null,
    address_delivery text not null,
    avatar text not null
);
