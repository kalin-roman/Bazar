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
