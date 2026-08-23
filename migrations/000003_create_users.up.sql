create table users(
    id serial primary key,
    full_name text not null,
    email text not null,
    age integer not null,
    address_delivery text not null,
    avatar text not null
);
