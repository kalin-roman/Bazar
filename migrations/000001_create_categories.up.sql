create table categories(
    id serial primary key,
    name text not null,
    slug text unique not null,
    image_url text not null
);
