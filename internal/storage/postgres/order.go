package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kalin-roman/Bazar/internal/order"
)

type OrderRepository struct {
	ConnectionPool *pgxpool.Pool
}

var _ order.Repository = (*OrderRepository)(nil)

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{ConnectionPool: pool}
}

// itemsForOrder fetches the line items belonging to a single order.
func (r *OrderRepository) itemsForOrder(ctx context.Context, orderID int64) ([]order.OrderItem, error) {
	rows, err := r.ConnectionPool.Query(ctx,
		"select listing_id, price_cents, quantity from order_items where order_id = $1",
		orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("list order items: %w", err)
	}
	defer rows.Close()

	var items []order.OrderItem
	for rows.Next() {
		var item order.OrderItem
		if err := rows.Scan(&item.ListingID, &item.Price, &item.Quantity); err != nil {
			return nil, fmt.Errorf("list order items: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list order items: %w", err)
	}

	return items, nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id int64) (order.Order, error) {
	row := r.ConnectionPool.QueryRow(ctx, "select id, user_id from orders where id = $1", id)

	var o order.Order
	if err := row.Scan(&o.ID, &o.UserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return order.Order{}, order.ErrNotFound
		}
		return order.Order{}, fmt.Errorf("get order by id: %w", err)
	}

	items, err := r.itemsForOrder(ctx, o.ID)
	if err != nil {
		return order.Order{}, err
	}
	o.Items = items

	return o, nil
}

func (r *OrderRepository) List(ctx context.Context) ([]order.Order, error) {
	rows, err := r.ConnectionPool.Query(ctx, "select id, user_id from orders")
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var orders []order.Order
	for rows.Next() {
		var o order.Order
		if err := rows.Scan(&o.ID, &o.UserID); err != nil {
			return nil, fmt.Errorf("list orders: %w", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}

	// Each order's items live in a separate table, so a second query
	// is needed per order to assemble the full Order value.
	for i := range orders {
		items, err := r.itemsForOrder(ctx, orders[i].ID)
		if err != nil {
			return nil, err
		}
		orders[i].Items = items
	}

	return orders, nil
}

func (r *OrderRepository) Create(ctx context.Context, o order.Order) (order.Order, error) {
	// Inserting an order means writing to two tables (orders, then
	// one row per item into order_items). A transaction ensures
	// that if any insert fails partway through, everything written
	// so far in this Create call is undone, rather than leaving a
	// half-written order with no items (or missing items) behind.
	tx, err := r.ConnectionPool.Begin(ctx)
	if err != nil {
		return order.Order{}, fmt.Errorf("create order: begin transaction: %w", err)
	}
	// Rollback is a no-op if the transaction was already committed,
	// so it's safe to defer unconditionally.
	defer tx.Rollback(ctx)

	if err := tx.QueryRow(ctx,
		"insert into orders (user_id) values ($1) returning id",
		o.UserID,
	).Scan(&o.ID); err != nil {
		return order.Order{}, fmt.Errorf("create order: %w", err)
	}

	for _, item := range o.Items {
		_, err := tx.Exec(ctx,
			"insert into order_items (order_id, listing_id, price_cents, quantity) values ($1, $2, $3, $4)",
			o.ID, item.ListingID, item.Price, item.Quantity,
		)
		if err != nil {
			return order.Order{}, fmt.Errorf("create order item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return order.Order{}, fmt.Errorf("create order: commit: %w", err)
	}

	return o, nil
}
