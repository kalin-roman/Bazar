package order

type Order struct {
	ID     int64
	UserID int64
	Items  []OrderItem
}

type OrderItem struct {
	ListingID int64
	Price     int64
	Quantity  int64
}
