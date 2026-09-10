-- Renaming, not recreating: RENAME TO/RENAME COLUMN preserve the
-- tables' data and automatically update any foreign key constraints
-- that reference them (order_items -> products FK stays intact,
-- just pointing at the new names). The sequence/constraint/index
-- renames below don't change any behavior — they just keep the
-- underlying object names honest instead of leaving "listing"
-- artifacts behind that RENAME TO/RENAME COLUMN don't touch.
alter table listings rename to products;
alter table listing_images rename to product_images;
alter table product_images rename column listing_id to product_id;
alter table order_items rename column listing_id to product_id;

alter sequence listings_id_seq rename to products_id_seq;
alter table products rename constraint listings_pkey to products_pkey;
alter table products rename constraint listings_slug_key to products_slug_key;
alter table products rename constraint listings_category_id_fkey to products_category_id_fkey;

alter sequence listing_images_id_seq rename to product_images_id_seq;
alter table product_images rename constraint listing_images_pkey to product_images_pkey;
alter table product_images rename constraint listing_images_listing_id_fkey to product_images_product_id_fkey;

alter table order_items rename constraint order_items_listing_id_fkey to order_items_product_id_fkey;
