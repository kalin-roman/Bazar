-- Reverse of up.sql, in reverse order.
alter table order_items rename constraint order_items_product_id_fkey to order_items_listing_id_fkey;

alter table product_images rename constraint product_images_product_id_fkey to listing_images_listing_id_fkey;
alter table product_images rename constraint product_images_pkey to listing_images_pkey;
alter sequence product_images_id_seq rename to listing_images_id_seq;

alter table products rename constraint products_category_id_fkey to listings_category_id_fkey;
alter table products rename constraint products_slug_key to listings_slug_key;
alter table products rename constraint products_pkey to listings_pkey;
alter sequence products_id_seq rename to listings_id_seq;

alter table order_items rename column product_id to listing_id;
alter table product_images rename column product_id to listing_id;
alter table product_images rename to listing_images;
alter table products rename to listings;
