package order

import "time"

type Order struct {
	ID        int64
	UserID    string
	Status    string
	CreatedAt time.Time
	Items     []OrderItem
}

type OrderItem struct {
	ProductID  int64
	PriceCents int64
	Quantity   int64
}
