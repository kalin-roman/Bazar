create table orders(
    id serial primary key,
    user_id integer not null references users(id)
);

create table order_items(
    id serial primary key,
    order_id integer not null references orders(id),
    listing_id integer not null references listings(id),
    price_cents integer not null,
    quantity integer not null
);
