-- The frontend's order-history UI shows a status badge and a date,
-- but until now orders carried neither on the backend. Adding real
-- columns rather than faking these client-side, since a status that
-- never actually changes would be misleading, not just incomplete.
alter table orders add column status text not null default 'Pending';
alter table orders add column created_at timestamptz not null default now();
